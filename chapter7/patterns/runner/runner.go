// Example is provided with help by Gabriel Aszalos.
// Package runner manages the running and lifetime of a process.
// runner包管理处理任务的运行和生命周期
package runner

import (
	"errors"
	"os"
	"os/signal"
	"time"
)

// Runner runs a set of tasks within a given timeout and can be
// shut down on an operating system interrupt.
// Runner 在给定的超时时间内执行一组任务
// 并且在操作系统发送中断信号时结束这些任务
type Runner struct {
	// interrupt channel reports a signal from the
	// operating system.
	// interrupt 通道报告从操作系统发送的信号
	interrupt chan os.Signal

	// complete channel reports that processing is done.
	// complete 通道报告处理任务已经完成
	complete chan error

	// timeout reports that time has run out.
	// timeout 报告处理任务已经超时
	timeout <-chan time.Time

	// tasks holds a set of functions that are executed
	// synchronously in index order.
	// tasks 持有一组以索引顺序依次执行的函数
	tasks []func(int)
}

// ErrTimeout is returned when a value is received on the timeout channel.
// ErrTimeout 会在任务执行超时时返回
var ErrTimeout = errors.New("received timeout")

// ErrInterrupt is returned when an event from the OS is received.
// ErrInterrupt 会在接收到操作系统的事件时返回
var ErrInterrupt = errors.New("received interrupt")

// New returns a new ready-to-use Runner.
// New 返回一个新的准备使用的Runner
func New(d time.Duration) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   time.After(d),
	}
}

// Add attaches tasks to the Runner. A task is a function that
// takes an int ID.
// ADD 将一个任务附加到Runner上。这个任务是一个接收一个int类型的ID作为参数的函数
func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

// Start runs all tasks and monitors channel events.
// Start 执行所有任务，并监视通道函数
func (r *Runner) Start() error {
	// We want to receive all interrupt based signals.
	// 我们希望接收所有中断信号
	signal.Notify(r.interrupt, os.Interrupt)

	// Run the different tasks on a different goroutine.
	// 用不同的goroutine执行不同的任务
	go func() {
		r.complete <- r.run()
	}()

	select {
	// 当任务处理完成时发出的信号
	// Signaled when processing is done.
	case err := <-r.complete:
		return err

	// Signaled when we run out of time.
	// 当任务处理程序运行超时时发出的信号
	case <-r.timeout:
		return ErrTimeout
	}
}

// run executes each registered task.
// run 执行每一个已注册的任务
func (r *Runner) run() error {
	for id, task := range r.tasks {
		// Check for an interrupt signal from the OS.
		// 检测操作系统的中断信号
		if r.gotInterrupt() {
			return ErrInterrupt
		}

		// Execute the registered task.
		// 执行已注册的任务
		task(id)
	}

	return nil
}

// gotInterrupt verifies if the interrupt signal has been issued.
// gotInterrupt 验证是否接收到中断信号
func (r *Runner) gotInterrupt() bool {
	select {
	// 当中断事件被触发时发出的信号
	// Signaled when an interrupt event is sent.
	case <-r.interrupt:
		// 停止接收后续的任何信号
		// Stop receiving any further signals.
		signal.Stop(r.interrupt)
		return true

	// Continue running as normal.
	// 继续正常运行
	default:
		return false
	}
}
