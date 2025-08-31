package k8sresources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func CreateorUpdateResource(ctx context.Context, c client.Client, scheme *runtime.Scheme, obj client.Object, log logr.Logger) error {
	// resourceLog := log.WithValues(
	// 	"name", obj.GetName(),
	// 	"namespace", obj.GetNamespace(),
	// )
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
