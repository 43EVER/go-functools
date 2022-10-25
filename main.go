package gofunctools

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

func padding(src []byte, blocksize int) []byte {
	n := len(src)
	padnum := blocksize - n%blocksize
	pad := bytes.Repeat([]byte{byte(padnum)}, padnum)
	dst := append(src, pad...)
	return dst
}

func unpadding(src []byte) []byte {
	n := len(src)
	unpadnum := int(src[n-1])
	dst := src[:n-unpadnum]
	return dst
}

// EncryptDES DES 加密
func EncryptDES(src []byte, key []byte) []byte {
	block, _ := des.NewCipher(key)
	src = padding(src, block.BlockSize())
	blockmode := cipher.NewCBCEncrypter(block, key)
	blockmode.CryptBlocks(src, src)
	return src
}

// DecryptDES DES 解密
func DecryptDES(src []byte, key []byte) []byte {
	block, _ := des.NewCipher(key)
	blockmode := cipher.NewCBCDecrypter(block, key)
	blockmode.CryptBlocks(src, src)
	src = unpadding(src)
	return src
}

func Map[T any, U any](data []T, fun func(T) U) []U {
	result := make([]U, 0)
	for _, item := range data {
		result = append(result, fun(item))
	}
	return result
}

func Filter[T any](data []T, fun func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range data {
		if fun(item) {
			result = append(result, item)
		}
	}
	return result
}

func Reduce[T any, U any](data []T, fun func(U, T) U, init U) U {
	result := init
	for _, item := range data {
		result = fun(result, item)
	}
	return result
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Float interface {
	~float32 | ~float64
}

type Ordered interface {
	Integer | Float | ~string
}

func Greater[T Ordered](a, b T) bool {
	return a > b
}

func Less[T Ordered](a, b T) bool {
	return a < b
}

func Max[T any](lessFunc func(a, b T) bool, data ...T) T {
	return Reduce(data, func(prev, cur T) T {
		if lessFunc(prev, cur) {
			return cur
		}
		return prev
	}, data[0])
}

func Min[T any](lessFunc func(a, b T) bool, data ...T) T {
	return Max(func(a, b T) bool { return !lessFunc(a, b) }, data...)
}

// TmpPtr 字面值变指针
func TmpPtr[T any](data T) *T {
	return &data
}

// FindIf 在 data 里找key，找到返回(下标, true)，没找到返回(0, false)
func FindIf[T any, U any](key T, data []U, predicate func(T, U) bool) (int, bool) {
	for idx, item := range data {
		if predicate(key, item) {
			return idx, true
		}
	}
	return 0, false
}

// Find 使用 == 作为谓词 的 FindIf
func Find[T comparable, U any](key T, data []U, keyFunc func(U) T) (int, bool) {
	return FindIf(key, data, func(t T, u U) bool {
		return t == keyFunc(u)
	})
}

// Identity 接到什么返什么
func Identity[T any](value T) T {
	return value
}

func Foldl[T any, U any](init U, fun func(U, T) U, data ...T) U {
	if len(data) == 0 {
		return init
	}
	return fun(Foldl(init, fun, data[1:]...), data[0])
}

func FoldR[T any, U any](init U, fun func(U, T) U, data ...T) U {
	if len(data) == 0 {
		return init
	}
	return FoldR(fun(init, data[0]), fun, data[1:]...)
}

func Compose[T any, U any, V any](fun1 func(T) U, fun2 func(U) V) func(T) V {
	return func(t T) V {
		return fun2(fun1(t))
	}
}

func IsSliceEq[T any, U any](srcs []T, targets []U, eqFun func(src T, tgt U) bool) bool {
	if len(srcs) != len(targets) {
		return false
	}

	return len(Filter(srcs, func(src T) bool {
		_, ok := FindIf(src, targets, eqFun)
		return !ok
	})) == 0
}

func BindFirst[T any, U any](f func(T) U, a1 T) func() U {
	return func() U { return f(a1) }
}

func Bind2First[T1 any, T2 any, U any](f func(T1, T2) U, a1 T1) func(T2) U {
	return func(a2 T2) U {
		return f(a1, a2)
	}
}

func ReverseParamter[T1 any, T2 any, U any](f func(T1, T2) U) func(T2, T1) U {
	return func(t2 T2, t1 T1) U {
		return f(t1, t2)
	}
}

func ReverseParamterWithError[T1 any, T2 any, U any](f func(T1, T2) (U, error)) func(T2, T1) (U, error) {
	return func(t2 T2, t1 T1) (U, error) {
		return f(t1, t2)
	}
}

var IntToString func(int64) string = Bind2First(ReverseParamter(strconv.FormatInt), 10)

// StringToInt 有 error 直接 panic
func StringToInt(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err.Error())
	}
	return i
}

func StructToMapByTag(value interface{}, tag string) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("ToMap only accepts structs; got %T", v))
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		name := fi.Name
		if tag != "" {
			if fi.Tag.Get(tag) != "" {
				name = fi.Tag.Get(tag)
			} else {
				continue
			}
		}

		if (fi.Type.Kind() != reflect.Slice && fi.Type.Kind() != reflect.Array && fi.Type.Kind() != reflect.Map) || (v.Field(i).Len() == 0) {
			if v.Field(i).Kind() == reflect.Struct || (v.Field(i).Kind() == reflect.Pointer && v.Field(i).Elem().Kind() == reflect.Struct) {
				result[name] = StructToMapByTag(v.Field(i).Interface(), tag)
			} else {
				result[name] = v.Field(i).Interface()
			}
		} else if fi.Type.Kind() == reflect.Map {
			tmp := map[string]interface{}{}
			for _, key := range v.Field(i).MapKeys() {
				value := v.Field(i).MapIndex(key)
				if _, ok := Find(value.Kind(), []reflect.Kind{
					reflect.Array, reflect.Struct, reflect.Slice, reflect.Map,
				}, Identity[reflect.Kind]); !ok {
					tmp[key.String()] = value.Interface()
				} else {
					tmp[key.String()] = StructToMapByTag(value.Interface(), tag)
				}
			}
			result[name] = tmp
		} else {
			tmp := make([]interface{}, 0)
			idx := 0
			for idx < v.Field(i).Len() {
				tmp = append(tmp, StructToMapByTag(v.Field(i).Index(idx).Interface(), tag))
				idx += 1
			}
			result[name] = tmp
		}

	}
	return result
}

// StructToMap 将结构体转为 map
func StructToMap[T any](s T) map[string]interface{} {
	return Bind2First(ReverseParamter(StructToMapByTag), "")(s)
}

// 深拷贝，string slice 也会拷贝一份新的
func CopyURLValues(src url.Values) url.Values {
	result := url.Values{}
	for key, value := range src {
		result[key] = make([]string, 0)
		result[key] = append(result[key], value...)
	}
	return result
}

// Unique Slice 去重，O(N) 的复杂度
func Unique[T comparable](data []T) []T {
	if len(data) < 1 {
		return []T{}
	}

	resultMap := make(map[T]bool)
	for _, item := range data {
		resultMap[item] = true
	}
	keys := make([]T, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}
	return keys
}

// UniqueWithEqFun 对 slice 去重，O(N^2) 的复杂度，数据长度 > 1000 时慎用
func UniqueWithEqFun[T any](data []T, eqFun func(T, T) bool) []T {
	if len(data) < 2 {
		return data
	}

	result := make([]T, 0)
	for _, item := range data {
		ok := true
		for _, alreadyInItem := range result {
			if eqFun(item, alreadyInItem) {
				ok = false
				break
			}
		}
		if ok {
			result = append(result, item)
		}
	}
	return result
}

// AndSlice 对两个 Slice 取交集，时间复杂度为O(max(N, M))，N为a的长度，M为b的长度
func AndSlice[T comparable](a []T, b []T) []T {
	resultMap := make(map[T]bool)
	bMap := make(map[T]bool)
	for _, item := range b {
		bMap[item] = true
	}

	for _, item := range a {
		if _, inBMap := bMap[item]; inBMap {
			resultMap[item] = true
		}
	}

	keys := make([]T, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}
	return keys
}

// SubSlice 两个 Slice 求差
func SubSlice[T comparable](a []T, b []T) []T {
	resultMap := make(map[T]bool)
	bMap := make(map[T]bool)
	for _, item := range b {
		bMap[item] = true
	}

	for _, item := range a {
		if _, inBMap := bMap[item]; !inBMap {
			resultMap[item] = true
		}
	}

	keys := make([]T, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}
	return keys
}

// PhoneMasking 电话脱敏，前后保留两位，其他全部省略
func PhoneMasking(phone string) string {
	str := []rune(phone)
	if len(str) < 5 {
		return phone
	}
	for m := range str {
		if m > 2 && m < len(str)-2 {
			str[m] = '*'
		}
	}
	return string(str)
}

// GenGetFieldFunc D 数据类型 R 返回值类型 只支持 struct 和 map，不支持嵌套
func GenGetFieldFunc[D, R any](field interface{}) func(D) R {
	return func(d D) R {
		v := reflect.ValueOf(d)
		// 一直拿，直至拿到非 pointer 类型
		for k := v.Kind(); k == reflect.Pointer; k = v.Kind() {
			v = v.Elem()
		}
		switch v.Kind() {
		case reflect.Struct:
			return v.FieldByName(field.(string)).Interface().(R)
		case reflect.Map:
			return v.MapIndex(reflect.ValueOf(field)).Interface().(R)
		default:
			panic("not supported type")
		}
	}
}

// RWLockDataWrapper 包了一层读写锁的变量
// 别传指针或按引用
type RWLockDataWrapper[T any] struct {
	lock sync.RWMutex
	data T
}

func (w *RWLockDataWrapper[T]) Read() T {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.data
}

func (w *RWLockDataWrapper[T]) Write(data T) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.data = data
}

func PtrDeref[T any](t *T) T {
	return *t
}

type Tuple[T, U any] struct {
	First  T
	Second U
}

func Zip[T, U any](s1 []T, s2 []U) []Tuple[T, U] {
	i := 0
	j := 0
	result := []Tuple[T, U]{}
	for i < len(s1) && j < len(s2) {
		result = append(result, Tuple[T, U]{
			First:  s1[i],
			Second: s2[j],
		})
		i += 1
		j += 1
	}
	return result
}

func ConvertSameType[R any](data interface{}, typeOf R) (R, bool) {
	result, ok := data.(R)
	return result, ok
}

func InterfaceToType[T any](v interface{}, typeof T) T {
	return v.(T)
}

func WrapWithLabel(data []string, leftLabel string, rightLabel string) string {
	return strings.Join(
		Map(data, func(item string) string {
			return leftLabel + item + rightLabel
		}),
		"",
	)
}

// ParseFormReq 从 req 中提取
func ParseFormReq[T any](value url.Values, parsedReq *T) error {
	decoder := form.NewDecoder()
	if err := decoder.Decode(&parsedReq, value); err != nil {
		return fmt.Errorf("params parse failed")
	}

	// struct 才需要 validate
	if reflect.TypeOf(parsedReq).Elem().Kind() != reflect.Struct {
		return nil
	}
	validate := validator.New()
	if err := validate.Struct(parsedReq); err != nil {
		return fmt.Errorf("params validate failed")
	}

	return nil
}
