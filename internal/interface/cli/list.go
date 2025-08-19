package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"version-list/internal/application"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有已安装的Go版本",
	Long:  `列出所有已安装的Go版本，并标记当前使用的版本`,
	Run: func(cmd *cobra.Command, args []string) {
		appService, err := application.NewVersionAppService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "初始化应用服务失败: %s", err)
			os.Exit(1)
		}

		versions, err := appService.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取版本列表失败: %s", err)
			os.Exit(1)
		}

		if len(versions) == 0 {
			fmt.Println("没有安装任何Go版本")
			return
		}

		// 使用tabwriter格式化输出
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "版本	路径	状态")
		for _, v := range versions {
			status := ""
			if v.IsActive {
				status = "当前使用"
			}
			fmt.Fprintf(w, "%s	%s	%s \r\n", v.Version, v.Path, status)
		}
		w.Flush()
	},
}
