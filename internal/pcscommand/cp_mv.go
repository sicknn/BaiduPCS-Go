package pcscommand

import (
	"fmt"
	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs"
	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs/pcserror"
	"path"
	"strings"
)

// RunCopy 执行 批量拷贝文件/目录
func RunCopy(paths ...string) {
	runCpMvOp("copy", paths...)
}

// RunMove 执行 批量 重命名/移动 文件/目录
func RunMove(paths ...string) {
	runCpMvOp("move", paths...)
}

func runCpMvOp(op string, paths ...string) {
	err := cpmvPathValid(paths...) // 检查路径的有效性, 目前只是判断数量
	if err != nil {
		fmt.Printf("%s path error, %s\n", op, err)
		return
	}

	from, to := cpmvParsePath(paths...) // 分割

	from, err = matchPathByShellPattern(from...)
	if err != nil {
		fmt.Println(err)
		return
	}
	to = GetActiveUser().PathJoin(to)

	// 尝试匹配
	if strings.ContainsAny(to, baidupcs.ShellPatternCharacters) {
		tos, _ := matchPathByShellPattern(to)

		switch len(tos) {
		case 0:
			// do nothing
		case 1:
			to = tos[0]
		default:
			fmt.Printf("Target directory has %d matches, please check your wildcard pattern\n", len(tos))
			return
		}
	}

	pcs := GetBaiduPCS()
	toInfo, pcsError := pcs.FilesDirectoriesMeta(to)
	switch {
	case toInfo != nil && toInfo.Path != path.Clean(to):
		fallthrough
	case pcsError != nil && pcsError.GetErrType() == pcserror.ErrTypeRemoteError:
		// 判断路径是否存在
		// 如果不存在, 则为重命名或同目录拷贝操作

		// 如果 from 数不是1, 则意义不明确.
		if len(from) != 1 {
			fmt.Println(err)
			return
		}

		if op == "copy" { // 拷贝
			err = pcs.Copy(&baidupcs.CpMvJSON{
				From: from[0],
				To:   to,
			})
			if err != nil {
				fmt.Println(err)
				fmt.Println("File/directory copy failed:")
				fmt.Printf("%s <-> %s\n", from[0], to)
				return
			}
			fmt.Println("File/directory copy succeeded:")
			fmt.Printf("%s <-> %s\n", from[0], to)
		} else { // 重命名
			err = pcs.Rename(from[0], path.Clean(to))
			if err != nil {
				fmt.Println(err)
				fmt.Println("Rename failed:")
				fmt.Printf("%s -> %s\n", from[0], to)
				return
			}
			fmt.Println("Rename succeeded:")
			fmt.Printf("%s -> %s\n", from[0], to)
		}
		return
	case pcsError != nil && pcsError.GetErrType() != pcserror.ErrTypeRemoteError:
		fmt.Println(pcsError)
		return
	}

	if !toInfo.Isdir {
		fmt.Printf("Target %s is not a directory, operation failed\n", toInfo.Path)
		return
	}

	cj := new(baidupcs.CpMvListJSON)
	cj.List = make([]*baidupcs.CpMvJSON, len(from))
	for k := range from {
		cj.List[k] = &baidupcs.CpMvJSON{
			From: from[k],
			To:   path.Clean(to + baidupcs.PathSeparator + path.Base(from[k])),
		}
	}

	switch op {
	case "copy":
		err = pcs.Copy(cj.List...)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Operation failed, the following files/directories failed to copy:")
			fmt.Println(cj)
			return
		}
		fmt.Println("Operation succeeded, copied files/directories:")
		fmt.Println(cj)
	case "move":
		err = pcs.Move(cj.List...)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Operation failed, the following files/directories failed to move:")
			fmt.Println(cj)
			return
		}
		fmt.Println("Operation succeeded, moved files/directories:")
		fmt.Println(cj)
	default:
		panic("Unknown operation:" + op)
	}
	return
}

// cpmvPathValid 检查路径的有效性
func cpmvPathValid(paths ...string) (err error) {
	if len(paths) <= 1 {
		return fmt.Errorf("incomplete arguments")
	}

	return nil
}

// cpmvParsePath 解析路径
func cpmvParsePath(paths ...string) (from []string, to string) {
	if len(paths) == 0 {
		return nil, ""
	}
	from = paths[:len(paths)-1]
	to = paths[len(paths)-1]
	return
}
