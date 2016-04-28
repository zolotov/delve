package memory

const cacheEnabled = true

type ReadWriter interface {
	Read(uint64, int) ([]byte, error)
	Write(uint64, []byte) (int, error)
	Swap(uint64, []byte) ([]byte, error)
}

type Memory struct {
	tid int
}

func (m *Memory) Read(addr uint64, size int) ([]byte, error) {
	return read(m.tid, addr, size)
}

func (m *Memory) Write(addr uint64, data []byte) (int, error) {
	return write(m.tid, addr, data)
}

func (m *Memory) Swap(addr uint64, data []byte) ([]byte, error) {
	originalData, err := read(m.tid, addr, len(data))
	if err != nil {
		return nil, err
	}
	if _, err := write(m.tid, addr, data); err != nil {
		return nil, err
	}
	return originalData, nil
}

type memCache struct {
	cacheAddr uint64
	cache     []byte
	mem       ReadWriter
}

func (m *memCache) contains(addr uint64, size int) bool {
	return addr >= m.cacheAddr && (addr+uint64(size)) <= (m.cacheAddr+uint64(len(m.cache)))
}

func (m *memCache) Read(addr uint64, size int) (data []byte, err error) {
	if m.contains(addr, size) {
		d := make([]byte, size)
		copy(d, m.cache[addr-m.cacheAddr:])
		return d, nil
	}

	return m.mem.Read(addr, size)
}

func (m *memCache) Write(addr uint64, data []byte) (written int, err error) {
	return m.mem.Write(addr, data)
}

func (m *memCache) Swap(addr uint64, data []byte) ([]byte, error) {
	return m.mem.Swap(addr, data)
}

func cacheMemory(mem ReadWriter, addr uint64, size int) ReadWriter {
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
