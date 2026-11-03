//go:build linux && with_gvisor

package core

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	vpn "github.com/wanan9999/veepin/l2tp"
)

// Runs only inside the CI job's isolated network namespace. No default route,
// host firewall or system DNS is changed. The server itself needs no OS TUN.
func TestL2TPFullPipeline(t *testing.T) {
	if raceEnabled {
		t.Skip("starting a Box trips sing-box NetworkManager races")
	}
	if os.Getenv("SUI_L2TP_INTEGRATION") != "1" {
		t.Skip("requires isolated Linux namespace and /dev/net/tun")
	}
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "alice-exit") }))
	defer a.Close()
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "bob-exit") }))
	defer b.Close()
	port := func(s *httptest.Server) int { return s.Listener.Addr().(*net.TCPAddr).Port }
	c := NewCore()
	cfg := fmt.Sprintf(`{
  "log":{"level":"error"},
  "inbounds":[{"type":"l2tp","tag":"vpn","listen":"127.0.0.1","listen_port":500,"public_ip":"127.0.0.1","psk":"test-only-psk","pool":"10.20.0.0/24","users":[{"name":"alice","password":"test-alice"},{"name":"bob","password":"test-bob"}]}],
  "outbounds":[{"type":"direct","tag":"exit-a"},{"type":"direct","tag":"exit-b"}],
  "dns":{"servers":[{"type":"hosts","tag":"test-dns","predefined":{"l2tp.test":["192.0.2.123"]}}]},
  "route":{"rules":[
   {"port":53,"action":"hijack-dns"},
   {"auth_user":"alice","action":"route","outbound":"exit-a","override_address":"127.0.0.1","override_port":%d},
   {"auth_user":"bob","action":"route","outbound":"exit-b","override_address":"127.0.0.1","override_port":%d},
   {"action":"reject"}
  ]}
 }`, port(a), port(b))
	if err := c.Start([]byte(cfg)); err != nil {
		t.Fatal(err)
	}
	defer c.Stop()
	ip := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("ip", args...).CombinedOutput(); err != nil {
			t.Fatalf("ip %v: %s: %v", args, out, err)
		}
	}
	for _, user := range []string{"alice", "bob"} {
		t.Run(user, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			session, result, err := vpn.Dial(ctx, vpn.Config{Server: "127.0.0.1", PSK: "test-only-psk", Username: user, Password: "test-" + user, TUNName: "suitest0"})
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			ip("addr", "add", result.AssignedIP.String()+"/32", "dev", result.TUNName)
			ip("link", "set", result.TUNName, "up")
			ip("route", "add", "198.18.0.1/32", "dev", result.TUNName, "src", result.AssignedIP.String())
			defer exec.Command("ip", "route", "del", "198.18.0.1/32", "dev", result.TUNName).Run()
			ip("route", "add", result.DNS[0].String()+"/32", "dev", result.TUNName)
			defer exec.Command("ip", "route", "del", result.DNS[0].String()+"/32", "dev", result.TUNName).Run()
			transport := &http.Transport{DisableKeepAlives: true}
			defer transport.CloseIdleConnections()
			client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
			response, err := client.Get("http://198.18.0.1/")
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil || string(body) != user+"-exit" {
				t.Fatalf("wrong user route: %q %v", body, err)
			}
			for _, network := range []string{"udp", "tcp"} {
				query := new(dns.Msg)
				query.SetQuestion("l2tp.test.", dns.TypeA)
				answer, _, err := (&dns.Client{Net: network, Timeout: 5 * time.Second}).Exchange(query, net.JoinHostPort(result.DNS[0].String(), "53"))
				if err != nil {
					t.Fatalf("%s DNS: %v", network, err)
				}
				if len(answer.Answer) != 1 || answer.Answer[0].(*dns.A).A.String() != "192.0.2.123" {
					t.Fatalf("%s DNS wrong answer: %v", network, answer)
				}
			}
			if user == "alice" {
				access, _ := c.GetInstance().Inbound().Get("vpn")
				type dnsChannel struct {
					client *dns.Client
					conn   *dns.Conn
				}
				var channels []dnsChannel
				query := new(dns.Msg)
				query.SetQuestion("l2tp.test.", dns.TypeA)
				for _, network := range []string{"tcp", "udp"} {
					dnsClient := &dns.Client{Net: network, Timeout: 5 * time.Second}
					dnsConn, err := dnsClient.Dial(net.JoinHostPort(result.DNS[0].String(), "53"))
					if err != nil {
						t.Fatal(err)
					}
					defer dnsConn.Close()
					if _, _, err = dnsClient.ExchangeWithConn(query, dnsConn); err != nil {
						t.Fatal(err)
					}
					channels = append(channels, dnsChannel{dnsClient, dnsConn})
				}
				apply := func(raw string) {
					t.Helper()
					update, err := c.PreparePolicy([]byte(raw))
					if err != nil {
						t.Fatal(err)
					}
					defer update.Abort()
					if err := update.Commit(); err != nil {
						t.Fatal(err)
					}
				}
				updatedDNS := strings.ReplaceAll(cfg, "192.0.2.123", "192.0.2.124")
				apply(updatedDNS)
				for _, channel := range channels {
					answer, _, err := channel.client.ExchangeWithConn(query, channel.conn)
					if err != nil || len(answer.Answer) != 1 || answer.Answer[0].(*dns.A).A.String() != "192.0.2.124" {
						t.Fatalf("existing %s DNS connection did not use new policy: %v %v", channel.client.Net, answer, err)
					}
				}
				updatedRoute := strings.ReplaceAll(updatedDNS, fmt.Sprintf(`"override_port":%d`, port(a)), fmt.Sprintf(`"override_port":%d`, port(b)))
				apply(updatedRoute)
				response, err := client.Get("http://198.18.0.1/")
				if err != nil {
					t.Fatal(err)
				}
				body, err := io.ReadAll(response.Body)
				response.Body.Close()
				if err != nil || string(body) != "bob-exit" {
					t.Fatalf("route change required a redial: %q %v", body, err)
				}
				stillAccess, _ := c.GetInstance().Inbound().Get("vpn")
				if stillAccess != access {
					t.Fatal("policy update recreated L2TP listener")
				}
				apply(cfg)
			}
			if n := c.KickUserSessions(user); n != 1 {
				t.Fatalf("kicked %d sessions", n)
			}
			if response, err := client.Get("http://198.18.0.1/"); err == nil {
				response.Body.Close()
				t.Fatal("kicked user still routed")
			}
		})
	}
}
