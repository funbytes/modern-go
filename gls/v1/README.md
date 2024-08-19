# goroutine local storage

Thanks https://github.com/huandu/go-tls for original idea

- get current goroutine id (unsafe.Pointer)
- goroutine local storage
- shard to 32-slot, more effective in Multi-Core CPU 

