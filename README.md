# ttlmap
Efficient concurrent map with TTL support. See the
[documentation](https://pkg.go.dev/github.com/job79/ttlmap)
for more information.

```go
ttlmap := New[string, string](time.Hour, time.Minute)
ttlmap.Store("key", "value")

val, ok := ttlmap.Load("key", "value")
val, ok := ttlmap.LoadOrStore("key", "value")
val, ok := ttlmap.LoadAndDelete("key")
ttlmap.Delete("key")
ttlmap.Range(func(key string, value string) bool {
	fmt.Println(key)
	return true
})
```
