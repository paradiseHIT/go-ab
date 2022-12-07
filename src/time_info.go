package main

import "time"

var startTime = time.Now()

// now returns time.Duration using stdlib time
func now() time.Duration { return time.Since(startTime) }
