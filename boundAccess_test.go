package rtree

//
//import (
//	"testing"
//)
//
//func BenchmarkDirectAccess(b *testing.B) {
//	i := newNode()
//	rawAccess(i, b.N)
//}
//
//func BenchmarkInterfaceMethod(b *testing.B) {
//	i := newNode()
//	interAccess(i, b.N)
//}
//
//func BenchmarkKnownType(b *testing.B) {
//	i := newNode()
//	knownType(i, b.N)
//}
//
//func BenchmarkTypeSwitch(b *testing.B) {
//	i := newNode()
//	checkAccess(i, b.N)
//}
//
//func rawAccess(i *node, n int) {
//	for k := 0; k < n; k++ {
//		_ = i.bounds
//	}
//}
//
//func interAccess(any Item, n int) {
//	for k := 0; k < n; k++ {
//		_ = any.Bounds()
//	}
//}
//
//func knownType(any Item, n int) {
//	for k := 0; k < n; k++ {
//		_ = any.(*node).bounds
//	}
//}
//
//func checkAccess(any Item, n int) {
//	for k := 0; k < n; k++ {
//		if it, ok := any.(*node); ok {
//			_ = it.bounds
//		} else {
//			_ = any.Bounds()
//		}
//	}
//}
