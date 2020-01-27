package cachet

import (
	"net"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/miekg/dns"
)

type DNSAnswer struct {
	Regex  string
	regexp *regexp.Regexp
	Exact  string
}

type DNSMonitor struct {
	AbstractMonitor `mapstructure:",squash"`

	// IP:port format or blank to use system defined DNS
	DNS string

	// A(default), AAAA, MX, ...
	Question string
	question uint16

	Answers []DNSAnswer
}

func (monitor *DNSMonitor) Validate() []string {
	errs := monitor.AbstractMonitor.Validate()

	if len(monitor.DNS) == 0 {
		config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
		if len(config.Servers) > 0 {
			monitor.DNS = net.JoinHostPort(config.Servers[0], config.Port)
		}
	}

	if len(monitor.DNS) == 0 {
		monitor.DNS = "8.8.8.8:53"
	}

	if len(monitor.Question) == 0 {
		monitor.Question = "A"
	}
	monitor.Question = strings.ToUpper(monitor.Question)

	monitor.question = findDNSType(monitor.Question)
	if monitor.question == 0 {
		errs = append(errs, "Could not look up DNS question type")
	}

	for i, a := range monitor.Answers {
		if len(a.Regex) > 0 {
			monitor.Answers[i].regexp, _ = regexp.Compile(a.Regex)
		}
	}

	return errs
}

func (monitor *DNSMonitor) test() bool {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(monitor.Target), monitor.question)
	m.RecursionDesired = true

	c := new(dns.Client)
	r, _, err := c.Exchange(m, monitor.DNS)
	if err != nil {
		logrus.Warnf("DNS error: %v", err)
		return false
	}

	if r.Rcode != dns.RcodeSuccess {
		return false
	}

	for _, check := range monitor.Answers {
		found := false
		for _, answer := range r.Answer {
			found = matchAnswer(answer, check)
			if found {
				break
			}
		}

		if !found {
			logrus.Warnf("DNS check failed: %v. Not found in any of %v", check, r.Answer)
			return false
		}
	}

	return true
}

func findDNSType(t string) uint16 {
	for rr, strType := range dns.TypeToString {
		if t == strType {
			return rr
		}
	}

	return 0
}

func matchAnswer(answer dns.RR, check DNSAnswer) bool {
	fields := []string{}
	for i := 0; i < dns.NumField(answer); i++ {
		fields = append(fields, dns.Field(answer, i+1))
	}

	str := strings.Join(fields, " ")

	if check.regexp != nil {
		return check.regexp.Match([]byte(str))
	}

	return str == check.Exact
}
