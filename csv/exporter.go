package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"

	"threatintel-feed-wizard/api"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

var csvHeader = []string{
	"id", "type", "value",
	"name", "description", "indicator_types",
	"pattern", "pattern_type", "pattern_version",
	"created", "modified", "valid_from", "valid_until",
	"confidence", "labels",
	"kill_chain_phases", "created_by_ref",
}

// stixPattern matches STIX 2.1 comparison expressions.
var stixPattern = regexp.MustCompile(`\[([a-zA-Z0-9_-]+):(\S+)\s*=\s*'([^']*)'`)

var iocTypeMap = map[string]string{
	"domain-name:value":             "domain",
	"ipv4-addr:value":               "ip",
	"ipv6-addr:value":               "ipv6",
	"url:value":                     "url",
	"email-addr:value":              "email",
	"file:hashes.'SHA-256'":         "sha256",
	"file:hashes.'SHA-1'":           "sha1",
	"file:hashes.'MD5'":             "md5",
	"file:hashes.SHA-256":           "sha256",
	"file:hashes.SHA-1":             "sha1",
	"file:hashes.MD5":               "md5",
	"file:name":                     "filename",
	"windows-registry-key:key":      "regkey",
	"process:name":                  "process",
	"network-traffic:dst_ref.value": "ip",
	"network-traffic:src_ref.value": "ip",
}

// ParsePattern extracts the IoC type and value from a STIX pattern.
func ParsePattern(pattern string) (iocType, iocValue string) {
	m := stixPattern.FindStringSubmatch(pattern)
	if m == nil {
		return "unknown", pattern
	}
	key := m[1] + ":" + m[2]
	if friendly, ok := iocTypeMap[key]; ok {
		return friendly, m[3]
	}
	return m[1], m[3]
}

func FormatIndicatorTypes(types []string) string {
	return strings.Join(types, ";")
}

func FormatLabels(labels []string) string {
	return strings.Join(labels, ";")
}

func FormatKillChainPhases(phases []api.KillChainPhase) string {
	parts := make([]string, len(phases))
	for i, p := range phases {
		parts[i] = p.KillChainName + ":" + p.PhaseName
	}
	return strings.Join(parts, ";")
}

func formatConfidence(c *int) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%d", *c)
}

// WriteCSV writes indicators to w in CSV format, prefixed with a UTF-8 BOM.
func WriteCSV(w io.Writer, indicators []api.Indicator) error {
	if _, err := w.Write(utf8BOM); err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	if err := cw.Write(csvHeader); err != nil {
		return err
	}

	for _, ind := range indicators {
		iocType, iocValue := ParsePattern(ind.Pattern)
		record := []string{
			ind.ID,
			iocType,
			iocValue,
			ind.Name,
			ind.Description,
			FormatIndicatorTypes(ind.IndicatorTypes),
			ind.Pattern,
			ind.PatternType,
			ind.PatternVersion,
			ind.Created,
			ind.Modified,
			ind.ValidFrom,
			ind.ValidUntil,
			formatConfidence(ind.Confidence),
			FormatLabels(ind.Labels),
			FormatKillChainPhases(ind.KillChainPhases),
			ind.CreatedByRef,
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
