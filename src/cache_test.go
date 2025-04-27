package src

import (
	"fmt"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	c := NewMyCache()
	c.Set("wang", "lc")
	fmt.Println(c.Get("wang"))
	c.Set("wang", "test")
	fmt.Println(c.Get("wang"))

}

func TestSetExp(t *testing.T) {
	c := NewMyCache()
	fmt.Println("永久有效")
	c.SetWithExpiration("wang", "lc", -1)
	time.Sleep(5 * time.Second)
	fmt.Println(c.Get("wang"))

	fmt.Println("有效期3s")
	c.SetWithExpiration("wang", "lc", 3)
	fmt.Println(c.Get("wang"))
	time.Sleep(5 * time.Second)
	fmt.Println(c.Get("wang"))

}

func TestCleanup(t *testing.T) {
	c := NewMyCache()
	c.SetWithExpiration("wang", "lc", 4)
	fmt.Println(c.Count())
	time.Sleep(6 * time.Second)
	fmt.Println(c.Count())
}

func TestSetNX(t *testing.T) {
	c := NewMyCache()
	c.Set("wang", "lc")
	result := c.SetNX("wang", "test")
	fmt.Println(result)

	c.Delete("wang")
	result1 := c.SetNX("wang", "test")
	fmt.Println(result1)

}
