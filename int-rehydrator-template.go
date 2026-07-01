package hydration

import "fmt"

func getHollowedGoContent(targetShortName string) string {
	return fmt.Sprintf(`package %[1]s
 
import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
	"unsafe"

	quic "sov.fleet/quic-go"
)
 
type jobResult struct {
	value string
	err   error
}
 
type Job struct {
	Prompt string
	Opts   []RefineOption
	Result chan jobResult
}
 
type RefineOptions struct {
	Chained  bool
	CallerID string
	Model    string
}
 
type RefineOption func(*RefineOptions)
 
func WithChained(chained bool) RefineOption {
	return func(o *RefineOptions) { o.Chained = chained }
}
 
func WithCallerID(callerID string) RefineOption {
	return func(o *RefineOptions) { o.CallerID = callerID }
}
 
func WithModel(model string) RefineOption {
	return func(o *RefineOptions) { o.Model = model }
}
 
type Conduit interface {
	Refine(prompt string, opts ...RefineOption) (string, error)
	Close() error
}

type ClientRequest struct {
	Action   string   "json:\"action\""
	Prompt   string   "json:\"prompt,omitempty\""
	Domain   string   "json:\"domain,omitempty\""
	Chained  bool     "json:\"chained,omitempty\""
	CallerID string   "json:\"caller_id,omitempty\""
	Model    string   "json:\"model,omitempty\""
}

type ClientResponse struct {
	Result string "json:\"result,omitempty\""
	Error  string "json:\"error,omitempty\""
}

func bytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

func marshalRequest(req *ClientRequest) []byte {
	domainBytes := []byte(req.Domain)
	promptBytes := []byte(req.Prompt)
	callerBytes := []byte(req.CallerID)
	modelBytes := []byte(req.Model)

	headerLen := 12
	totalLen := headerLen + len(domainBytes) + len(promptBytes) + len(callerBytes) + len(modelBytes)
	buf := make([]byte, totalLen)

	buf[0] = 0x01
	if req.Action == "login" {
		buf[0] = 0x02
	}
	if req.Chained {
		buf[1] = 1
	} else {
		buf[1] = 0
	}

	domLen := len(domainBytes)
	buf[2] = byte(domLen >> 8)
	buf[3] = byte(domLen)

	prLen := len(promptBytes)
	buf[4] = byte(prLen >> 24)
	buf[5] = byte(prLen >> 16)
	buf[6] = byte(prLen >> 8)
	buf[7] = byte(prLen)

	cLen := len(callerBytes)
	buf[8] = byte(cLen >> 8)
	buf[9] = byte(cLen)

	mLen := len(modelBytes)
	buf[10] = byte(mLen >> 8)
	buf[11] = byte(mLen)

	offset := headerLen
	copy(buf[offset:], domainBytes)
	offset += len(domainBytes)
	copy(buf[offset:], promptBytes)
	offset += len(promptBytes)
	copy(buf[offset:], callerBytes)
	offset += len(callerBytes)
	copy(buf[offset:], modelBytes)

	return buf
}

func unmarshalResponse(stream io.Reader) (*ClientResponse, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(stream, header); err != nil {
		return nil, err
	}

	resLen := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])
	errLen := uint32(header[4])<<24 | uint32(header[5])<<16 | uint32(header[6])<<8 | uint32(header[7])

	payload := make([]byte, resLen+errLen)
	if _, err := io.ReadFull(stream, payload); err != nil {
		return nil, err
	}

	resBytes := payload[:resLen]
	errBytes := payload[resLen:]

	return &ClientResponse{
		Result: bytesToString(resBytes),
		Error:  bytesToString(errBytes),
	}, nil
}
 
type remoteConduit struct {
	domain string
}

var (
	gConn *quic.Conn
	gConnMu sync.Mutex
)

func getOrDialConnection(ctx context.Context) (*quic.Conn, error) {
	gConnMu.Lock()
	defer gConnMu.Unlock()
	if gConn != nil {
		return gConn, nil
	}

	binaryPath, err := findDaemonPath()
	if err != nil {
		return nil, err
	}

	var conn *quic.Conn
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"wraith-protocol"},
	}
	quicConf := &quic.Config{
		MaxIdleTimeout:                 0,
		InitialStreamReceiveWindow:     1024 * 1024 * 32,
		MaxStreamReceiveWindow:         1024 * 1024 * 64,
		InitialConnectionReceiveWindow: 1024 * 1024 * 64,
		MaxConnectionReceiveWindow:     1024 * 1024 * 128,
		MaxIncomingStreams:             100000,
		DisablePathMTUDiscovery:        true,
		KeepAlivePeriod:                0,
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4242")
	if err != nil {
		return nil, err
	}

	for attempt := 0; attempt < 5; attempt++ {
		rawConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0})
		if err == nil {
			c, dialErr := quic.Dial(ctx, rawConn, udpAddr, tlsConf, quicConf)
			if dialErr == nil {
				conn = c
				break
			}
			rawConn.Close()
		}

		if attempt == 0 {
			cmd := exec.Command(binaryPath, "-daemon", "-port", "4242")
			cmd.Stdout = nil
			cmd.Stderr = nil
			if startErr := cmd.Start(); startErr != nil {
				return nil, fmt.Errorf("failed to start daemon: %%v", startErr)
			}
		}
		time.Sleep(150 * time.Millisecond)
	}

	if conn == nil {
		return nil, fmt.Errorf("failed to connect to daemon at 127.0.0.1:4242")
	}

	authStream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		conn.CloseWithError(0, "failed to open auth stream")
		return nil, fmt.Errorf("failed to open auth stream: %%v", err)
	}
	defer authStream.Close()

	licensePath := filepath.Join(filepath.Dir(binaryPath), "..", "config", "license.json")
	licData, err := os.ReadFile(licensePath)
	if err != nil {
		conn.CloseWithError(0, "failed to read license")
		return nil, fmt.Errorf("failed to read license config: %%v", err)
	}

	type LicenseInfo struct {
		Email string "json:\"email\""
		Token string "json:\"token\""
	}
	var lic LicenseInfo
	if err := json.Unmarshal(licData, &lic); err != nil {
		conn.CloseWithError(0, "failed to parse license")
		return nil, fmt.Errorf("failed to parse license JSON: %%v", err)
	}

	type AuthRequest struct {
		Email string "json:\"email\""
		Token string "json:\"token\""
	}
	type AuthResponse struct {
		Status string "json:\"status\""
		Error  string "json:\"error\""
	}

	if err := json.NewEncoder(authStream).Encode(AuthRequest{Email: lic.Email, Token: lic.Token}); err != nil {
		conn.CloseWithError(0, "failed to write auth")
		return nil, fmt.Errorf("failed to write auth payload: %%v", err)
	}

	var authRes AuthResponse
	if err := json.NewDecoder(authStream).Decode(&authRes); err != nil {
		conn.CloseWithError(0, "failed to read auth response")
		return nil, fmt.Errorf("failed to read auth response: %%v", err)
	}

	if authRes.Status != "OK" {
		conn.CloseWithError(0, "unauthorized")
		return nil, fmt.Errorf("daemon authorization failed: %%s", authRes.Error)
	}

	gConn = conn
	return gConn, nil
}

func getStream(ctx context.Context) (*quic.Conn, *quic.Stream, error) {
	for i := 0; i < 2; i++ {
		c, err := getOrDialConnection(ctx)
		if err != nil {
			return nil, nil, err
		}
		stream, err := c.OpenStreamSync(ctx)
		if err == nil {
			return c, stream, nil
		}
		
		gConnMu.Lock()
		if gConn == c {
			gConn.CloseWithError(0, "stale")
			gConn = nil
		}
		gConnMu.Unlock()
	}
	return nil, nil, fmt.Errorf("failed to acquire active stream")
}
 
func (rc *remoteConduit) Refine(prompt string, opts ...RefineOption) (string, error) {
	ctx := context.Background()
	_, stream, err := getStream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	ro := &RefineOptions{}
	for _, o := range opts {
		o(ro)
	}

	req := ClientRequest{
		Action:   "refine",
		Prompt:   prompt,
		Domain:   rc.domain,
		Chained:  ro.Chained,
		CallerID: ro.CallerID,
		Model:    ro.Model,
	}

	if _, err := stream.Write(marshalRequest(&req)); err != nil {
		return "", err
	}

	res, err := unmarshalResponse(stream)
	if err != nil {
		return "", err
	}

	if res.Error != "" {
		return "", fmt.Errorf("remote error: %%s", res.Error)
	}

	return res.Result, nil
}
 
func (rc *remoteConduit) Close() error {
	return nil
}
 
type DOMProfile struct {
	InputSelectors  []string
	SubmitSelectors []string
	StopSelectors   []string
	SuccessPatterns []string
	InjectMode      string
	WaitMs          int
}
 
var GeminiProfile = DOMProfile{
	InputSelectors:  []string{"textarea"},
	SubmitSelectors: []string{"button"},
}
 
var ClaudeProfile = DOMProfile{
	InputSelectors:  []string{"textarea"},
	SubmitSelectors: []string{"button"},
}
 
func New(ctx context.Context, domain string, profile DOMProfile) (Conduit, error) {
	return &remoteConduit{domain: domain}, nil
}
 
func InteractiveLogin(ctx context.Context, targetURL string, profile DOMProfile) error {
	_, stream, err := getStream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	req := ClientRequest{
		Action: "login",
		Domain: targetURL,
	}

	if _, err := stream.Write(marshalRequest(&req)); err != nil {
		return err
	}

	res, err := unmarshalResponse(stream)
	if err != nil {
		return err
	}

	if res.Error != "" {
		return fmt.Errorf("remote login error: %%s", res.Error)
	}

	return nil
}
 
func findDaemonPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := cwd
	for {
		if _, err := os.Stat(filepath.Join(current, ".gitroot")); err == nil {
			return filepath.Join(current, ".nvelwraith", "bin", "wraithd.exe"), nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf(".gitroot not found")
}
`, targetShortName)
}
