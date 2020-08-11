package resolver

import (
	"fmt"
	"strings"

	. "nickns/resolver/esxi"
)

/*
// Manage cache existing
var cache QueryCaches
var hasCache = map[string]bool{}

// Manage cache detail
type QueryCache struct {
	Fqdn   string
	IpAddr string
	Expire time.Time
}
type QueryCaches []QueryCache
*/

// Config for esxi hosts info
func SetEsxiConfigPath(hostFilePath string) {
	EsxiNodeConfPath = hostFilePath
}

// Resolve type 'A' record
func ResolveRecordTypeA(hostname string) string {
	/* cache hit
	if hasCache[fqdn] {
		log.Printf("[CacheHit] %s\n", fqdn)
		for _, vm := range cache {
			// todo: check cache-expire -> del cache from set and array
			if vm.Fqdn == strings.Split(fqdn, ".")[0] {
				return vm.IpAddr
			}
		}
	}
	*/
	for _, vm := range GetAllVmIdName() {
		// log.Println(vm.Name, fqdn)
		if vm.Name == hostname {
			return GetVmIp(vm)
		}
	}
	return "" // UnHit
}

// Resolve type 'A' record
func ResolveRecordTypePTR(ptrAddr string) string {
	/*
		hasCache[vmFqdn] = true
		cache = append(cache, QueryCache{
			Fqdn:   vmFqdn,
			IpAddr: vmIp,
			Expire: time.Now(),
		})
	*/
	slice := strings.Split(ptrAddr, ".")
	ipAddr := fmt.Sprintf("%s.%s.%s.%s", slice[3], slice[2], slice[1], slice[0])

	if hostname := GetVmIpName(ipAddr); hostname == "" {
		return ""
	} else {
		return hostname
	}
}
