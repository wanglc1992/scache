package src

import "time"

func (c *MyCache) Run() {

	go func() {
		ticker := time.NewTicker(c.cleanInterval)

		for {
			select {
			case <-ticker.C:
				c.Cleanup()
			case <-c.stopCh:
				return
			}
		}
	}()
}

func (c *MyCache) _stop() {
	close(c.stopCh)
}
