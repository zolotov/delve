package memory

const cacheEnabled = true

type ReadWriter interface {
	Read(addr uintptr, size int) ([]byte, error)
	Write(addr uintptr, data []byte) (int, error)
}

type memCache struct {
	cacheAddr uintptr
	cache     []byte
	mem       ReadWriter
}

func (m *memCache) contains(addr uintptr, size int) bool {
	return addr >= m.cacheAddr && (addr+uintptr(size)) <= (m.cacheAddr+uintptr(len(m.cache)))
}

func (m *memCache) Read(addr uintptr, size int) (data []byte, err error) {
	if m.contains(addr, size) {
		d := make([]byte, size)
		copy(d, m.cache[addr-m.cacheAddr:])
		return d, nil
	}

	return m.mem.Read(addr, size)
}

func (m *memCache) Write(addr uintptr, data []byte) (written int, err error) {
	return m.mem.Write(addr, data)
}

func cacheMemory(mem ReadWriter, addr uintptr, size int) ReadWriter {
	if !cacheEnabled {
		return mem
	}
	if size <= 0 {
		return mem
	}
	if cacheMem, isCache := mem.(*memCache); isCache {
		if cacheMem.contains(addr, size) {
			return mem
		} else {
			cache, err := cacheMem.mem.Read(addr, size)
			if err != nil {
				return mem
			}
			return &memCache{addr, cache, mem}
		}
	}
	cache, err := mem.Read(addr, size)
	if err != nil {
		return mem
	}
	return &memCache{addr, cache, mem}
}
