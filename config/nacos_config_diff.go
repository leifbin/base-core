package config

import (
	"fmt"
	"log/slog"
	"reflect"
)

// CompareConfigs 比较两个对象是否不同，并打印差异（如果有）
// 该函数使用反射，可以支持任意结构体的差异打印，不局限于特定的配置模型。
func CompareConfigs(oldCfg, newCfg interface{}) bool {
	if oldCfg == nil {
		return true
	}

	if reflect.DeepEqual(oldCfg, newCfg) {
		return false
	}

	slog.Info("🔍 检测到配置内容变更:")
	printDiff(reflect.ValueOf(oldCfg), reflect.ValueOf(newCfg), "")
	return true
}

func printDiff(vOld, vNew reflect.Value, prefix string) {
	// 处理指针：如果其中一个是 nil，直接打印差异
	if vOld.Kind() == reflect.Ptr {
		if vOld.IsNil() {
			if !vNew.IsNil() {
				slog.Info(fmt.Sprintf("  - %s 变更:", prefix), "old", nil, "new", vNew.Elem().Interface())
			}
			return
		}
		if vNew.IsNil() {
			slog.Info(fmt.Sprintf("  - %s 变更:", prefix), "old", vOld.Elem().Interface(), "new", nil)
			return
		}
		vOld = vOld.Elem()
		vNew = vNew.Elem()
	}

	if vOld.Kind() != reflect.Struct || vNew.Kind() != reflect.Struct {
		if !reflect.DeepEqual(vOld.Interface(), vNew.Interface()) {
			slog.Info(fmt.Sprintf("  - %s 变更:", prefix), "old", vOld.Interface(), "new", vNew.Interface())
		}
		return
	}

	tOld := vOld.Type()
	for i := 0; i < vOld.NumField(); i++ {
		field := tOld.Field(i)
		fName := field.Name
		if prefix != "" {
			fName = prefix + "." + fName
		}

		fOld := vOld.Field(i)
		fNew := vNew.Field(i)

		if reflect.DeepEqual(fOld.Interface(), fNew.Interface()) {
			continue
		}

		if fOld.Kind() == reflect.Struct || (fOld.Kind() == reflect.Ptr && fOld.Elem().Kind() == reflect.Struct) {
			printDiff(fOld, fNew, fName)
		} else {
			slog.Info(fmt.Sprintf("  - %s 变更:", fName), "old", fOld.Interface(), "new", fNew.Interface())
		}
	}
}
