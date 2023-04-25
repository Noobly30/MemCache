package cache

import "time"

func TestCachOP(t *testing.T) {
	testData := []struct {
		key    string
		val    interface{}
		expire time.Duration
	}{
		{key:"widqnsadada",679,time.Second*10}
		{key:"asdawawojd",643,time.Second*10}
		{key:"jxviqem,zsad",235,time.Second*10}
	}
	c:=NewMemCache()
	c.SetMaxMemory("10MB")
	for _, item := range testData{
		c.Set(item.key, item.val, item.expire)
		val, ok := c.Get(item.key)
		if !ok{
			t.Error("运存加载失败")
		}
		
	}
}