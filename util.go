// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/07/2021

package reqtest

import (
	"net"
)

// getOutboundIP will get preferred ip of this machine
func getOutboundIP() string {
	if c, err := net.Dial("udp", "8.8.8.8:80"); err == nil {
		defer c.Close()
		loc := c.LocalAddr().(*net.UDPAddr)
		return loc.IP.String()	
	}
	return ""
}
