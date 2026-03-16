package securitylog

import (
	"log"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/requestid"
)

// Log writes one structured security event using the process logger.
func Log(c *gin.Context, event string, fields map[string]string) {
	parts := []string{
		"component=security",
		"event=" + sanitize(event),
		"request_id=" + sanitize(requestid.FromContext(c)),
		"client_ip=" + sanitize(c.ClientIP()),
	}

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts = append(parts, sanitize(key)+"="+sanitize(fields[key]))
	}

	log.Printf("%s", strings.Join(parts, " "))
}

func sanitize(value string) string {
	if value == "" {
		return "-"
	}

	value = strings.ReplaceAll(value, "\r", "_")
	value = strings.ReplaceAll(value, "\n", "_")
	value = strings.ReplaceAll(value, "\t", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}
