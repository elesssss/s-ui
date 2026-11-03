package core

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	mDNS "github.com/miekg/dns"
	"github.com/sagernet/sing-box/adapter"
	M "github.com/sagernet/sing/common/metadata"
)

func policyTestConfig(answer, timeout string) []byte {
	return []byte(fmt.Sprintf(`{"log":{"level":"error"},"outbounds":[{"type":"direct","tag":"a","connect_timeout":%q},{"type":"direct","tag":"b"}],"dns":{"servers":[{"type":"hosts","tag":"hosts","predefined":{"reload.test":[%q]}}]},"route":{"rules":[{"auth_user":"bob","action":"route","outbound":"b"}],"final":"a"}}`, timeout, answer))
}

func policyTestCore(t *testing.T) *Core {
	t.Helper()
	if raceEnabled {
		t.Skip("native NetworkManager startup has an upstream race; policy ownership is covered separately")
	}
	c := NewCore()
	if err := c.Start(policyTestConfig("192.0.2.1", "5s")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Stop() })
	return c
}

func policyTestConnection(t *testing.T, c *Core, user string) net.Conn {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = io.Copy(conn, conn)
	}()
	client, server := net.Pipe()
	t.Cleanup(func() { client.Close(); server.Close() })
	go c.GetInstance().policy.RouteConnectionEx(context.Background(), server, adapter.InboundContext{Inbound: "test", User: user, Source: M.ParseSocksaddr("192.0.2.10:1234"), Destination: M.SocksaddrFromNet(listener.Addr())}, nil)
	policyEcho(t, client)
	return client
}
func policyEcho(t *testing.T, conn net.Conn) {
	t.Helper()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	var response [4]byte
	if _, err := io.ReadFull(conn, response[:]); err != nil {
		t.Fatal(err)
	}
	if string(response[:]) != "ping" {
		t.Fatalf("echo: %q", response)
	}
	_ = conn.SetDeadline(time.Time{})
}
func policyDNSAnswer(t *testing.T, c *Core, want string) {
	t.Helper()
	query := new(mDNS.Msg)
	query.SetQuestion("reload.test.", mDNS.TypeA)
	d := &accessDNS{policy: c.GetInstance().policy}
	answer, err := d.Exchange(context.Background(), query, adapter.DNSQueryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(answer.Answer) != 1 || answer.Answer[0].(*mDNS.A).A.String() != want {
		t.Fatalf("DNS: %v", answer)
	}
}

func TestPolicyDNSReloadPreservesBusinessConnections(t *testing.T) {
	c := policyTestCore(t)
	box := c.GetInstance()
	conn := policyTestConnection(t, c, "alice")
	policyDNSAnswer(t, c, "192.0.2.1")
	update, err := c.PreparePolicy(policyTestConfig("192.0.2.2", "5s"))
	if err != nil {
		t.Fatal(err)
	}
	defer update.Abort()
	policyDNSAnswer(t, c, "192.0.2.1") // preparation has no externally visible change
	if err := update.Commit(); err != nil {
		t.Fatal(err)
	}
	if c.GetInstance() != box {
		t.Fatal("policy update rebuilt core/access")
	}
	policyDNSAnswer(t, c, "192.0.2.2")
	policyEcho(t, conn)
}

// A candidate's interface monitor can reset networking before publication.
// Exercise that callback directly so this also detects cross-generation
// ownership on machines with only a loopback interface.
func TestPolicyCandidateNetworkResetIsIsolated(t *testing.T) {
	c := policyTestCore(t)
	conn := policyTestConnection(t, c, "alice")
	for _, commit := range []bool{false, true} {
		update, err := c.PreparePolicy(policyTestConfig("192.0.2.2", "5s"))
		if err != nil {
			t.Fatal(err)
		}
		defer update.Abort()
		update.runtime.network.ResetNetwork(context.Background())
		policyEcho(t, conn)
		if commit {
			if err := update.Commit(); err != nil {
				t.Fatal(err)
			}
			fresh := policyTestConnection(t, c, "alice")
			update.runtime.network.ResetNetwork(context.Background())
			_ = fresh.SetReadDeadline(time.Now().Add(time.Second))
			if _, err := fresh.Read(make([]byte, 1)); err != io.EOF {
				t.Fatalf("network reset did not close its own generation: %v", err)
			}
		} else {
			update.Abort()
		}
		policyEcho(t, conn)
	}
}

func TestPolicyOutboundReloadClosesOnlyDependentConnections(t *testing.T) {
	c := policyTestCore(t)
	alice := policyTestConnection(t, c, "alice")
	bob := policyTestConnection(t, c, "bob")
	update, err := c.PreparePolicy(policyTestConfig("192.0.2.1", "6s"))
	if err != nil {
		t.Fatal(err)
	}
	defer update.Abort()
	if err := update.Commit(); err != nil {
		t.Fatal(err)
	}
	_ = alice.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := alice.Read(make([]byte, 1)); err != io.EOF {
		t.Fatalf("old outbound connection was not closed: %v", err)
	}
	policyEcho(t, bob)
	policyEcho(t, policyTestConnection(t, c, "alice"))
}

func TestPolicyPreparationFailureAndAbortKeepCurrent(t *testing.T) {
	c := policyTestCore(t)
	conn := policyTestConnection(t, c, "alice")
	bad := strings.Replace(string(policyTestConfig("192.0.2.2", "5s")), `"final":"a"`, `"final":"missing"`, 1)
	if update, err := c.PreparePolicy([]byte(bad)); err == nil {
		update.Abort()
		t.Fatal("invalid default outbound accepted")
	}
	badRule := strings.Replace(string(policyTestConfig("192.0.2.2", "5s")), `"outbound":"b"`, `"outbound":"missing"`, 1)
	if update, err := c.PreparePolicy([]byte(badRule)); err == nil {
		update.Abort()
		t.Fatal("dangling route outbound accepted")
	}
	policyDNSAnswer(t, c, "192.0.2.1")
	policyEcho(t, conn)
	update, err := c.PreparePolicy(policyTestConfig("192.0.2.2", "5s"))
	if err != nil {
		t.Fatal(err)
	}
	update.Abort()
	update.Abort()
	policyDNSAnswer(t, c, "192.0.2.1")
	policyEcho(t, conn)
}

func TestPolicyRulesForceReconnectWithoutRebuildingCore(t *testing.T) {
	c := policyTestCore(t)
	box := c.GetInstance()
	conn := policyTestConnection(t, c, "alice")
	next := strings.Replace(string(policyTestConfig("192.0.2.1", "5s")), `"final":"a"`, `"final":"b"`, 1)
	update, err := c.PreparePolicy([]byte(next))
	if err != nil {
		t.Fatal(err)
	}
	defer update.Abort()
	if err := update.Commit(); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := conn.Read(make([]byte, 1)); err != io.EOF {
		t.Fatalf("old rule connection was not closed: %v", err)
	}
	if box != c.GetInstance() {
		t.Fatal("access core replaced")
	}
	policyEcho(t, policyTestConnection(t, c, "alice"))
}

func TestPolicyAbortClosesCandidateScopeOnly(t *testing.T) {
	c := policyTestCore(t)
	conn := policyTestConnection(t, c, "alice")
	box := c.GetInstance()
	update, err := c.PreparePolicy(policyTestConfig("192.0.2.2", "5s"))
	if err != nil {
		t.Fatal(err)
	}
	scope := update.runtime.scope
	update.Abort()
	update.Abort()
	if scope.Context().Err() == nil || box.scope.Context().Err() != nil {
		t.Fatal("abort leaked the candidate scope or cancelled the access scope")
	}
	policyDNSAnswer(t, c, "192.0.2.1")
	policyEcho(t, conn)
}

func TestPolicyGlobalChangeRefusedWithoutDisconnect(t *testing.T) {
	c := policyTestCore(t)
	conn := policyTestConnection(t, c, "alice")
	raw := strings.Replace(string(policyTestConfig("192.0.2.2", "5s")), `"level":"error"`, `"level":"info"`, 1)
	if update, err := c.PreparePolicy([]byte(raw)); err == nil {
		update.Abort()
		t.Fatal("global settings changed during a policy update")
	}
	policyDNSAnswer(t, c, "192.0.2.1")
	policyEcho(t, conn)
}
