package config

import (
	"fmt"
	"log/slog"
	"reflect"
)

// DiffType 定义变更类型
type DiffType string

const (
	// DiffUpdate 表示配置项被修改
	DiffUpdate DiffType = "UPDATE"
	// DiffAdd 表示新增了配置项
	DiffAdd DiffType = "ADD"
	// DiffDelete 表示删除了配置项
	DiffDelete DiffType = "DELETE"
)

// ConfigDiff 描述配置中的一个具体变更项
type ConfigDiff struct {
	Path     string      // 变更的路径，例如 "Base.AppPort"
	OldValue interface{} // 旧值
	NewValue interface{} // 新值
	Type     DiffType    // 变更类型
}

// CompareConfigs 比较两个对象是否不同，并返回详细的差异列表
func CompareConfigs(oldCfg, newCfg interface{}) (bool, []ConfigDiff) {
	if oldCfg == nil {
		return true, []ConfigDiff{{Path: "root", NewValue: newCfg, Type: DiffAdd}}
	}

	if reflect.DeepEqual(oldCfg, newCfg) {
		return false, nil
	}

	var diffs []ConfigDiff
	diffs = getDiffs(reflect.ValueOf(oldCfg), reflect.ValueOf(newCfg), "", diffs)

	if len(diffs) > 0 {
		slog.Info("🔍 检测到配置内容变更:")
		for _, d := range diffs {
			slog.Info(fmt.Sprintf("  [%s] - %s:", d.Type, d.Path), "old", d.OldValue, "new", d.NewValue)
		}
	}

	return len(diffs) > 0, diffs
}

func getDiffs(vOld, vNew reflect.Value, prefix string, diffs []ConfigDiff) []ConfigDiff {
	// 处理指针
	if vOld.Kind() == reflect.Ptr {
		if vOld.IsNil() {
			if !vNew.IsNil() {
				diffs = append(diffs, ConfigDiff{Path: prefix, OldValue: nil, NewValue: vNew.Elem().Interface(), Type: DiffAdd})
			}
			return diffs
		}
		if vNew.IsNil() {
			diffs = append(diffs, ConfigDiff{Path: prefix, OldValue: vOld.Elem().Interface(), NewValue: nil, Type: DiffDelete})
			return diffs
		}
		vOld = vOld.Elem()
		vNew = vNew.Elem()
	}

	// 1. 如果是 Map
	if vOld.Kind() == reflect.Map && vNew.Kind() == reflect.Map {
		// 找出删除和更新的
		for _, key := range vOld.MapKeys() {
			oldVal := vOld.MapIndex(key)
			newVal := vNew.MapIndex(key)
			keyStr := fmt.Sprintf("%v", key.Interface())
			newPath := prefix
			if prefix != "" {
				newPath += "." + keyStr
			} else {
				newPath = keyStr
			}

			if !newVal.IsValid() {
				// 删除
				diffs = append(diffs, ConfigDiff{Path: newPath, OldValue: oldVal.Interface(), NewValue: nil, Type: DiffDelete})
			} else if !reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
				// 递归比较具体内容
				diffs = getDiffs(oldVal, newVal, newPath, diffs)
			}
		}
		// 找出新增的
		for _, key := range vNew.MapKeys() {
			if !vOld.MapIndex(key).IsValid() {
				keyStr := fmt.Sprintf("%v", key.Interface())
				newPath := prefix
				if prefix != "" {
					newPath += "." + keyStr
				} else {
					newPath = keyStr
				}
				diffs = append(diffs, ConfigDiff{Path: newPath, OldValue: nil, NewValue: vNew.MapIndex(key).Interface(), Type: DiffAdd})
			}
		}
		return diffs
	}

	// 2. 如果是 Slice/Array
	if (vOld.Kind() == reflect.Slice || vOld.Kind() == reflect.Array) &&
		(vNew.Kind() == reflect.Slice || vNew.Kind() == reflect.Array) {
		if !reflect.DeepEqual(vOld.Interface(), vNew.Interface()) {
			// 对于列表，我们暂且认为整个列表发生了变化（或者可以进一步做更精细的索引比较）
			diffs = append(diffs, ConfigDiff{Path: prefix, OldValue: vOld.Interface(), NewValue: vNew.Interface(), Type: DiffUpdate})
		}
		return diffs
	}

	// 3. 如果是结构体
	if vOld.Kind() == reflect.Struct && vNew.Kind() == reflect.Struct {
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
			diffs = getDiffs(fOld, fNew, fName, diffs)
		}
		return diffs
	}

	// 4. 其他基础类型比较
	if !reflect.DeepEqual(vOld.Interface(), vNew.Interface()) {
		diffs = append(diffs, ConfigDiff{Path: prefix, OldValue: vOld.Interface(), NewValue: vNew.Interface(), Type: DiffUpdate})
	}
	return diffs
}
