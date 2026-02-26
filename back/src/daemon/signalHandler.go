package daemon

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"zsxyww.com/wts/server"
)

// regExitSigs注册OS信号处理器，用于捕获各种中止信号并进行收尾工作。
func regExitSigs() {

	sigs := make(chan os.Signal, 1)

	// SIGINT:  Ctrl+C
	// SIGTERM: kill命令
	// SIGQUIT: Ctrl+\
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigs
		fmt.Printf("\n===== %s,roger that! =====\n", sig)

		err := runCleanup()
		if err != nil {
			os.Exit(1)
		}

		//unix-like systems' exit code convention
		s := sig.(syscall.Signal)
		os.Exit(128 + int(s))

	}()

}

func runCleanup() error {
	fmt.Println("\n===== Starting Cleanup Program =====")
	//TODO:数据库之类
	err := saveWXAccessToken()
	fmt.Println("\n===== End Cleanup Program =====")
	return err
}

func saveWXAccessToken() error {
	actok, err := server.WX.GetAccessToken()
	if err != nil {
		fmt.Println("Failed to get WeChat access token while saving:", err)
		return err
	}
	//TODO：保存到.dat文件
	_ = actok
	fmt.Println("WeChat access token saved successfully.")
	return nil
}
