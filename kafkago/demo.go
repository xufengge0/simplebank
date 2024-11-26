package kafkago

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

func Scanner() {
    scanner := bufio.NewScanner(os.Stdin)

    fmt.Print("请输入数字的数量: ")
	
    scanner.Scan() // 读取数字的数量
    n, _ := strconv.Atoi(scanner.Text())

    nums := make([]int, n)
    p := make([]int, n)

    fmt.Println("请输入数字:")
    for i := 0; i < n; i++ {
        scanner.Scan() // 读取每个数字
        nums[i], _ = strconv.Atoi(scanner.Text())
        if i == 0 {
            p[i] = nums[i]
        } else {
            p[i] = nums[i] + p[i-1]
        }
    }

    for {
        fmt.Print("请输入范围 (begin end): ")
        scanner.Scan() // 读取范围
        parts := strings.Fields(scanner.Text())
        if len(parts) < 2 {
            fmt.Println("输入不完整，请输入两个数字。")
            continue
        }
        
        begin, _ := strconv.Atoi(parts[0])
        end, _ := strconv.Atoi(parts[1])

        // 检查范围有效性
        if begin < 0 || end >= n || begin > end {
            fmt.Println("范围无效，请重新输入。")
            continue
        }

        // 计算结果
        var res int
        if begin == 0 {
            res = p[end] // 从 0 到 end 的和
        } else {
            res = p[end] - p[begin-1] // 从 begin 到 end 的和
        }

        fmt.Println("范围内的和是:", res)
    }
}

