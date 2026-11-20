package main

import (
	"context"
	"fmt"
	"time"
)

func take(ctx context.Context, jobs <-chan string) (string, error) {
	select {
	case job := <-jobs:
		return job, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := take(ctx, make(chan string))
	fmt.Println(err)
}
