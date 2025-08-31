package k8sresources

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	keydbv1 "github.com/rsingh0101/keydb-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenerateSecret(k *keydbv1.Keydb, scheme *runtime.Scheme, c client.Client) *corev1.Secret {
	labels := map[string]string{"apps": k.Name}
	secretName := k.Name + "-secret"

	// Default: use Spec.Password if provided
	setPassword := k.Spec.Password

	// Try to fetch existing secret to reuse password
	var existing corev1.Secret
	err := c.Get(context.TODO(), types.NamespacedName{
		Name:      secretName,
		Namespace: k.Namespace,
	}, &existing)

	if err == nil {
		if pw, ok := existing.Data["password"]; ok && len(pw) > 0 {
			setPassword = string(pw)
		}
	}

	// If still empty, generate a random one
	if setPassword == "" {
		pw, err := Generatepassword(16)
		if err != nil {
			// fallback: very unlikely, but don't crash reconciliation
			pw = "changeme"
		}
		setPassword = pw
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: k.Namespace,
			Labels:    labels,
		},
		Data: map[string][]byte{
			"password": []byte(setPassword),
		},
	}

	_ = ctrl.SetControllerReference(k, secret, scheme)
	return secret
}

func Generatepassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
