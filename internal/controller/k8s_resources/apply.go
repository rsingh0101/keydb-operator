package k8sresources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func checksumConfigMap(cm *corev1.ConfigMap) string {
	h := sha256.New()
	for k, v := range cm.Data {
		h.Write([]byte(k))
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func CreateorUpdateResource(ctx context.Context, c client.Client, scheme *runtime.Scheme, obj client.Object, log logr.Logger) error {
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	existing := obj.DeepCopyObject().(client.Object)
	if err := c.Get(ctx, key, existing); err != nil {
		if apierrors.IsNotFound(err) {
			if err := c.Create(ctx, obj); err != nil {
				log.Error(err, "failed to create resource")
				return err
			}
			PrintResourceLog(ctx, "Created", obj, scheme)
		} else {
			log.Error(err, "failed to fetch resource")
			return err
		}
	} else {
		if err := c.Update(ctx, obj); err != nil {
			log.Error(err, "failed to update resource")
			return err
		}
		PrintResourceLog(ctx, "Updated", obj, scheme)
		if obj.GetObjectKind().GroupVersionKind().Kind == "ConfigMap" {
			sts := &appsv1.StatefulSet{}
			stsKey := types.NamespacedName{
				Name:      strings.TrimSuffix(obj.GetName(), "-config"),
				Namespace: obj.GetNamespace(),
			}
			if err := c.Get(ctx, stsKey, sts); err != nil {
				log.Error(err, "failed to fetch statefulset")
				return err
			}
			if sts.Spec.Template.Annotations == nil {
				sts.Spec.Template.Annotations = make(map[string]string)

			}
			sts.Spec.Template.Annotations["restartedAt"] = time.Now().Format(time.RFC3339)
			if err := c.Update(ctx, sts); err != nil {
				log.Error(err, "failed to restart statefulset")
				return err
			}
			PrintResourceLog(ctx, "Restarted Statefulset successfully", sts, scheme)
		}
	}
	return nil
}

func PrintResourceLog(ctx context.Context, action string, obj client.Object, scheme *runtime.Scheme) {
	details := map[string]string{
		"name":      obj.GetName(),
		"namespace": obj.GetNamespace(),
	}
	d, _ := json.Marshal(details)

	// Detect the kind
	gvk, err := apiutil.GVKForObject(obj, scheme)
	kind := "Unknown"
	if err == nil {
		kind = gvk.Kind
	}

	// ANSI colors
	green := "\033[32m"
	blue := "\033[34m"
	yellow := "\033[33m"
	red := "\033[31m"
	reset := "\033[0m"

	// Emojis + colored action
	var colorAction string
	switch action {
	case "Created":
		colorAction = fmt.Sprintf("%s✅ %s%s", green, action, reset)
	case "Updated":
		colorAction = fmt.Sprintf("%s🔄 %s%s", blue, action, reset)
	case "Deleted":
		colorAction = fmt.Sprintf("%s🗑️ %s%s", red, action, reset)
	default:
		colorAction = fmt.Sprintf("%sℹ️ %s%s", yellow, action, reset)
	}

	fmt.Printf("%s %s %s\n", colorAction, kind, string(d))
}
