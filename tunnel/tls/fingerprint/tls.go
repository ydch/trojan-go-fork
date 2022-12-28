package fingerprint

import (
	"crypto/tls"
	
	"github.com/Potterli20/trojan-go-fork/common"
	"github.com/Potterli20/trojan-go-fork/log"
	. "github.com/Potterli20/utls-fork"
)


func ParseCipher(s []string) []uint16 {
	all := tls.CipherSuites()
	var result []uint16
	for _, p := range s {
		found := true
		for _, q := range all {
			if q.Name == p {
				result = append(result, q.ID)
				break
			}
			if !found {
				log.Warn("invalid cipher suite", p, "skipped")
			}
		}
	}
	return result
}
