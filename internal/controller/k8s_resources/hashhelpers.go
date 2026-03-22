// deterministic.go
package k8sresources

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func HashConfigMap(cm *corev1.ConfigMap) string {
	if cm == nil {
		return ""
	}
	keys := make([]string, 0, len(cm.Data))
	for k := range cm.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("\n")
		b.WriteString(cm.Data[k])
		b.WriteString("\n")
	}

	// BinaryData (rare for ConfigMap, but future-proof)
	bkeys := make([]string, 0, len(cm.BinaryData))
	for k := range cm.BinaryData {
		bkeys = append(bkeys, k)
	}
	sort.Strings(bkeys)
	for _, k := range bkeys {
		b.WriteString(k)
		b.WriteString("\n")
		b.Write(cm.BinaryData[k])
		b.WriteString("\n")
	}

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func HashSecret(sec *corev1.Secret) string {
	if sec == nil {
		return ""
	}
	keys := make([]string, 0, len(sec.Data))
	for k := range sec.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("\n")
		b.Write(sec.Data[k])
		b.WriteString("\n")
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
