package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	"github.com/peterh/liner"
	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcscommand"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsfunctions/pcsdownload"
	_ "github.com/qjfoidnh/BaiduPCS-Go/internal/pcsinit"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsupdate"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsliner"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsliner/args"
	"github.com/qjfoidnh/BaiduPCS-Go/pcstable"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/checksum"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/escaper"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/getip"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/pcstime"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsverbose"
	"github.com/urfave/cli"
)

const (
	// NameShortDisplayNum 
	NameShortDisplayNum = 16

	cryptoDescription = `
	Available methods <method>:
		aes-128-ctr, aes-192-ctr, aes-256-ctr,
		aes-128-cfb, aes-192-cfb, aes-256-cfb,
		aes-128-ofb, aes-192-ofb, aes-256-ofb.

	Key <key>:
		aes-128 requires a 16-byte key, aes-192 requires 24 bytes, aes-256 requires 32 bytes.
		If key length does not match, extra bytes are trimmed and missing bytes are padded with '\0'.

	GZIP <disable-gzip>:
		Enable GZIP before encryption and after decryption (enabled by default).
		Disabling GZIP removes integrity hints, so source files are kept on decrypt to avoid data loss.`
)

var (
	// Version 
	Version = "v4.0.0-stable"

	historyFilePath = filepath.Join(pcsconfig.GetConfigDir(), "pcs_command_history.txt")
	reloadFn        = func(c *cli.Context) error {
		err := pcsconfig.Config.Reload()
		if err != nil {
			fmt.Printf("Failed to reload config: %s\n", err)
		}
		return nil
	}
	saveFunc = func(c *cli.Context) error {
		err := pcsconfig.Config.Save()
		if err != nil {
			fmt.Printf("Failed to save config: %s\n", err)
		}
		return nil
	}

	isCli bool
)

func init() {
	pcsutil.ChWorkDir()

	err := pcsconfig.Config.Init()
	switch err {
	case nil:
	case pcsconfig.ErrConfigFileNoPermission, pcsconfig.ErrConfigContentsParseError:
		fmt.Fprintf(os.Stderr, "FATAL ERROR: config file error: %s\n", err)
		os.Exit(1)
	default:
		fmt.Printf("WARNING: config init error: %s\n", err)
	}
}

func main() {
	defer pcsconfig.Config.Close()

	app := cli.NewApp()
	app.Name = "BaiduPCS-Go"
	app.Version = Version
	app.Author = "qjfoidnh/BaiduPCS-Go: https://github.com/qjfoidnh/BaiduPCS-Go"
	app.Copyright = "(c) 2016-2020 iikira."
	app.Usage = "Baidu Netdisk CLI for " + runtime.GOOS + "/" + runtime.GOARCH
	app.Description = `BaiduPCS-Go is a command-line Baidu Netdisk client written in Go.
Use COMMANDS below to manage files, download/upload data, and automate netdisk tasks.

Project: https://github.com/qjfoidnh/BaiduPCS-Go
Releases: https://github.com/qjfoidnh/BaiduPCS-Go/releases
Issues: https://github.com/qjfoidnh/BaiduPCS-Go/issues
Email: qjfoidnh@126.com`

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "Enable debug logging",
			EnvVar:      pcsverbose.EnvVerbose,
			Destination: &pcsverbose.IsVerbose,
		},
	}
	app.Action = func(c *cli.Context) {
		if c.NArg() != 0 {
			fmt.Printf("Command not found: %s\nRun %s help for usage\n", c.Args().Get(0), app.Name)
			return
		}

		isCli = true
		pcsverbose.Verbosef("VERBOSE: this is a debug message\n\n")

		var (
			line = pcsliner.NewLiner()
			err  error
		)

		line.History, err = pcsliner.NewLineHistory(historyFilePath)
		if err != nil {
			fmt.Printf("Warning: failed to read command history file, %s\n", err)
		}

		line.ReadHistory()
		defer func() {
			line.DoWriteHistory()
			line.Close()
		}()

		// tab 
		line.State.SetCompleter(func(line string) (s []string) {
			var (
				lineArgs                   = args.Parse(line)
				numArgs                    = len(lineArgs)
				acceptCompleteFileCommands = []string{
					"cd", "cp", "download", "export", "locate", "ls", "meta", "mkdir", "mv", "rm", "setastoken", "share", "transfer", "tree", "upload",
				}
				closed = strings.LastIndex(line, " ") == len(line)-1
			)

			for _, cmd := range app.Commands {
				for _, name := range cmd.Names() {
					if !strings.HasPrefix(name, line) {
						continue
					}

					s = append(s, name+" ")
				}
			}

			switch numArgs {
			case 0:
				return
			case 1:
				if !closed {
					return
				}
			}

			thisCmd := app.Command(lineArgs[0])
			if thisCmd == nil {
				return
			}

			if !pcsutil.ContainsString(acceptCompleteFileCommands, thisCmd.FullName()) {
				return
			}

			var (
				activeUser  = pcsconfig.Config.ActiveUser()
				pcs         = pcsconfig.Config.ActiveUserBaiduPCS()
				runeFunc    = unicode.IsSpace
				pcsRuneFunc = func(r rune) bool {
					switch r {
					case '\'', '"':
						return true
					}
					return unicode.IsSpace(r)
				}
				targetPath string
			)

			if !closed {
				targetPath = lineArgs[numArgs-1]
				escaper.EscapeStringsByRuneFunc(lineArgs[:numArgs-1], runeFunc) // 
			} else {
				escaper.EscapeStringsByRuneFunc(lineArgs, runeFunc)
			}

			switch {
			case targetPath == "." || strings.HasSuffix(targetPath, "/."):
				s = append(s, line+"/")
				return
			case targetPath == ".." || strings.HasSuffix(targetPath, "/.."):
				s = append(s, line+"/")
				return
			}

			var (
				targetDir string
				isAbs     = path.IsAbs(targetPath)
				isDir     = strings.LastIndex(targetPath, "/") == len(targetPath)-1
			)

			if isAbs {
				targetDir = path.Dir(targetPath)
			} else {
				targetDir = path.Join(activeUser.Workdir, targetPath)
				if !isDir {
					targetDir = path.Dir(targetDir)
				}
			}
			files, err := pcs.CacheFilesDirectoriesList(targetDir, baidupcs.DefaultOrderOptions)
			if err != nil {
				return
			}

			// fmt.Println("-", targetDir, targetPath, "-")

			for _, file := range files {
				if file == nil {
					continue
				}

				var (
					appendLine string
				)

				// 
				if !closed {
					if !strings.HasPrefix(file.Path, path.Clean(path.Join(targetDir, path.Base(targetPath)))) {
						if path.Base(targetDir) == path.Base(targetPath) {
							appendLine = strings.Join(append(lineArgs[:numArgs-1], escaper.EscapeByRuneFunc(path.Join(targetPath, file.Filename), pcsRuneFunc)), " ")
							goto handle
						}
						// fmt.Println(file.Path, targetDir, targetPath)
						continue
					}
					// fmt.Println(path.Clean(path.Join(path.Dir(targetPath), file.Filename)), targetPath, file.Filename)
					appendLine = strings.Join(append(lineArgs[:numArgs-1], escaper.EscapeByRuneFunc(path.Clean(path.Join(path.Dir(targetPath), file.Filename)), pcsRuneFunc)), " ")
					goto handle
				}
				// 
				appendLine = strings.Join(append(lineArgs, escaper.EscapeByRuneFunc(file.Filename, pcsRuneFunc)), " ")
				goto handle

			handle:
				if file.Isdir {
					s = append(s, appendLine+"/")
					continue
				}
				s = append(s, appendLine+" ")
				continue
			}

			return
		})

		fmt.Printf("Tip: Use Up/Down arrows to switch command history.\n")
		fmt.Printf("Tip: Use Ctrl+A/E to move to command start/end.\n")
		fmt.Printf("Tip: Type help for usage.\n")

		for {
			var (
				prompt     string
				activeUser = pcsconfig.Config.ActiveUser()
			)

			if activeUser.Name != "" {
				// : BaiduPCS-Go:<> <ID>$
				// , 
				prompt = app.Name + ":" + converter.ShortDisplay(path.Base(activeUser.Workdir), NameShortDisplayNum) + " " + activeUser.Name + "$ "
			} else {
				// BaiduPCS-Go >
				prompt = app.Name + " > "
			}

			commandLine, err := line.State.Prompt(prompt)
			switch err {
			case liner.ErrPromptAborted:
				return
			case nil:
				// continue
			default:
				fmt.Println(err)
				return
			}

			line.State.AppendHistory(commandLine)

			cmdArgs := args.Parse(commandLine)
			if len(cmdArgs) == 0 {
				continue
			}

			s := []string{os.Args[0]}
			s = append(s, cmdArgs...)

			// 
			// , 
			line.Pause()
			c.App.Run(s)
			line.Resume()
		}
	}

	app.Commands = []cli.Command{
		{
			Name:     "run",
			Usage:    "Run system command",
			Category: "Other",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				cmd := exec.Command(c.Args().First(), c.Args().Tail()...)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					fmt.Println(err)
				}

				return nil
			},
		},
		{
			Name:  "env",
			Usage: "Show environment variables",
			Description: "See command usage and options for details.",
			Category: "Other",
			Action: func(c *cli.Context) error {
				envStr := "%s=\"%s\"\n"
				envVar, ok := os.LookupEnv(pcsverbose.EnvVerbose)
				if ok {
					fmt.Printf(envStr, pcsverbose.EnvVerbose, envVar)
				} else {
					fmt.Printf(envStr, pcsverbose.EnvVerbose, "0")
				}

				envVar, ok = os.LookupEnv(pcsconfig.EnvConfigDir)
				if ok {
					fmt.Printf(envStr, pcsconfig.EnvConfigDir, envVar)
				} else {
					fmt.Printf(envStr, pcsconfig.EnvConfigDir, pcsconfig.GetConfigDir())
				}

				return nil
			},
		},
		{
			Name:     "update",
			Usage:    "Check for updates",
			Category: "Other",
			Action: func(c *cli.Context) error {
				if c.IsSet("y") {
					if !c.Bool("y") {
						return nil
					}
				}
				pcsupdate.CheckUpdate(app.Version, c.Bool("y"))
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "y",
					Usage: "Confirm update",
				},
			},
		},
		{
			Name:  "login",
			Usage: "Login Baidu account",
			Description: "See command usage and options for details.",
			Category: "Baidu Account",
			Before:   reloadFn,
			After:    saveFunc,
			Action: func(c *cli.Context) error {
				var bduss, ptoken, stoken, cookies string
				if c.IsSet("cookies") {
					cookies = c.String("cookies")
				} else if c.IsSet("bduss") {
					bduss = c.String("bduss")
					ptoken = c.String("ptoken")
					stoken = c.String("stoken")
				} else if c.NArg() == 0 {
					var err error
					bduss, ptoken, stoken, cookies, err = pcscommand.RunLogin(c.String("username"), c.String("password"))
					if err != nil {
						fmt.Println(err)
						return err
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				baidu, err := pcsconfig.Config.SetupUserByBDUSS(bduss, ptoken, stoken, cookies)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				fmt.Println("Baidu account login succeeded:", baidu.Name)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username",
					Usage: "Baidu account username (phone/email/username)",
				},
				cli.StringFlag{
					Name:  "password",
					Usage: "Password for Baidu account username",
				},
				cli.StringFlag{
					Name:  "bduss",
					Usage: "Use Baidu BDUSS to login",
				},
				cli.StringFlag{
					Name:  "ptoken",
					Usage: "Baidu PTOKEN used with -bduss (optional)",
				},
				cli.StringFlag{
					Name:  "stoken",
					Usage: "Baidu STOKEN used with -bduss (optional, required for transfer)",
				},
				cli.StringFlag{
					Name:  "cookies",
					Usage: "Use Baidu cookies to login",
				},
			},
		},
		{
			Name:  "su",
			Usage: "Switch Baidu account",
			Description: "See command usage and options for details.",
			Category: "Baidu Account",
			Before:   reloadFn,
			After:    saveFunc,
			Action: func(c *cli.Context) error {
				if c.NArg() >= 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				numLogins := pcsconfig.Config.NumLogins()

				if numLogins == 0 {
					fmt.Printf("No Baidu account configured, cannot switch\n")
					return nil
				}

				var (
					inputData = c.Args().Get(0)
					uid       uint64
				)

				if c.NArg() == 1 {
					// 
					uid, _ = strconv.ParseUint(inputData, 10, 64)
				} else if c.NArg() == 0 {
					// 
					cli.HandleAction(app.Command("loglist").Action, c)

					//  index
					var index string
					fmt.Printf("Enter # index to switch account > ")
					_, err := fmt.Scanln(&index)
					if err != nil {
						return nil
					}

					if n, err := strconv.Atoi(index); err == nil && n >= 0 && n < numLogins {
						uid = pcsconfig.Config.BaiduUserList[n].UID
					} else {
						fmt.Printf("Failed to switch user, please check the # index\n")
						return nil
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
				}

				switchedUser, err := pcsconfig.Config.SwitchUser(&pcsconfig.BaiduBase{
					Name: inputData,
				})
				if err != nil {
					switchedUser, err = pcsconfig.Config.SwitchUser(&pcsconfig.BaiduBase{
						UID: uid,
					})
					if err != nil {
						fmt.Printf("Failed to switch user, %s\n", err)
						return nil
					}
				}

				fmt.Printf("Switched user: %s\n", switchedUser.Name)
				return nil
			},
		},
		{
			Name:        "logout",
			Usage:       "Logout Baidu account",
			Description: "Log out current Baidu account",
			Category:    "Baidu Account",
			Before:      reloadFn,
			After:       saveFunc,
			Action: func(c *cli.Context) error {
				if pcsconfig.Config.NumLogins() == 0 {
					fmt.Println("No Baidu account configured, cannot logout")
					return nil
				}

				var (
					confirm    string
					activeUser = pcsconfig.Config.ActiveUser()
				)

				if !c.Bool("y") {
					fmt.Printf("Confirm logout account: %s ? (y/n) > ", activeUser.Name)
					_, err := fmt.Scanln(&confirm)
					if err != nil || (confirm != "y" && confirm != "Y") {
						return err
					}
				}

				deletedUser, err := pcsconfig.Config.DeleteUser(&pcsconfig.BaiduBase{
					UID: activeUser.UID,
				})
				if err != nil {
					fmt.Printf("Failed to logout user %s, error: %s\n", activeUser.Name, err)
				}

				fmt.Printf("Logout succeeded, %s\n", deletedUser.Name)
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "y",
					Usage: "Confirm account logout",
				},
			},
		},
		{
			Name:        "loglist",
			Usage:       "List account list",
			Description: "List all logged-in Baidu accounts",
			Category:    "Baidu Account",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Println(pcsconfig.Config.BaiduUserList.String())
				return nil
			},
		},
		{
			Name:  "setastoken",
			Usage: "Set accessToken for current account",
			Description: "See command usage and options for details.",
			Category: "Baidu Account",
			Before:   reloadFn,
			After:    saveFunc,
			Action: func(c *cli.Context) error {
				activeUser := pcsconfig.Config.ActiveUser()
				if activeUser.UID == 0 {
					fmt.Println("Please login first")
					return nil
				}
				if c.NArg() >= 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				} else if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				activeUser.AccessToken = c.Args().Get(0)
				pcsconfig.Config.ActiveUserBaiduPCS().SetaccessToken(c.Args().Get(0))
				fmt.Printf("Current username: %s, accessToken set: %s\n", activeUser.Name, activeUser.AccessToken)
				return nil
			},
		},
		{
			Name:        "who",
			Usage:       "Get current account",
			Description: "Get current account information",
			Category:    "Baidu Account",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				activeUser := pcsconfig.Config.ActiveUser()
				fmt.Printf("Current account uid: %d, username: %s, gender: %s, age: %.1f\n", activeUser.UID, activeUser.Name, activeUser.Sex, activeUser.Age)
				return nil
			},
		},
		{
			Name:        "quota",
			Usage:       "Get cloud storage quota",
			Description: "Get total and used cloud storage space",
			Category:    "Baidu Netdisk",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				pcscommand.RunGetQuota()
				return nil
			},
		},
		{
			Name:     "cd",
			Category: "Baidu Netdisk",
			Usage:    "Change working directory",
			Description: "See command usage and options for details.",
			Before: reloadFn,
			After:  saveFunc,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunChangeDirectory(c.Args().Get(0), c.Bool("l"))

				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l",
					Usage: "List workdir contents after changing directory",
				},
			},
		},
		{
			Name:      "ls",
			Aliases:   []string{"l", "ll"},
			Usage:     "List directory",
			UsageText: app.Name + " ls <dir>",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				orderOptions := &baidupcs.OrderOptions{}
				switch {
				case c.IsSet("asc"):
					orderOptions.Order = baidupcs.OrderAsc
				case c.IsSet("desc"):
					orderOptions.Order = baidupcs.OrderDesc
				default:
					orderOptions.Order = baidupcs.OrderAsc
				}

				switch {
				case c.IsSet("time"):
					orderOptions.By = baidupcs.OrderByTime
				case c.IsSet("name"):
					orderOptions.By = baidupcs.OrderByName
				case c.IsSet("size"):
					orderOptions.By = baidupcs.OrderBySize
				default:
					orderOptions.By = baidupcs.OrderByName
				}

				pcscommand.RunLs(c.Args().Get(0), &pcscommand.LsOptions{
					Total: c.Bool("l") || c.Parent().Args().Get(0) == "ll",
				}, orderOptions)

				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l",
					Usage: "Detailed output",
				},
				cli.BoolFlag{
					Name:  "asc",
					Usage: "Sort ascending",
				},
				cli.BoolFlag{
					Name:  "desc",
					Usage: "Sort descending",
				},
				cli.BoolFlag{
					Name:  "time",
					Usage: "Sort by time",
				},
				cli.BoolFlag{
					Name:  "name",
					Usage: "Sort by name",
				},
				cli.BoolFlag{
					Name:  "size",
					Usage: "Sort by size",
				},
			},
		},
		{
			Name:      "search",
			Aliases:   []string{"s"},
			Usage:     "Search files",
			UsageText: app.Name + " search [-path=<directory>] [-r] <keyword>",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunSearch(c.String("path"), c.Args().Get(0), &pcscommand.SearchOptions{
					Total:   c.Bool("l"),
					Recurse: c.Bool("r"),
				})

				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l",
					Usage: "Detailed output",
				},
				cli.BoolFlag{
					Name:  "r",
					Usage: "Search recursively",
				},
				cli.StringFlag{
					Name:  "path",
					Usage: "directory to search",
					Value: ".",
				},
			},
		},
		{
			Name:      "tree",
			Aliases:   []string{"t"},
			Usage:     "List directory tree",
			UsageText: app.Name + " tree <dir>",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				pcscommand.RunTree(c.Args().Get(0), 0, &pcscommand.TreeOptions{
					Depth:    c.Int("depth"),
					ShowFsid: c.Bool("fsid"),
				})
				return nil
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "depth",
					Usage: "Tree depth",
					Value: -1,
				},
				cli.BoolFlag{
					Name:  "fsid",
					Usage: "Include fsid",
				},
			},
		},
		{
			Name:      "pwd",
			Usage:     "Print working directory",
			UsageText: app.Name + " pwd",
			Category:  "Baidu Netdisk",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Println(pcsconfig.Config.ActiveUser().Workdir)
				return nil
			},
		},
		{
			Name:        "meta",
			Usage:       "Get file/directory metadata",
			UsageText:   app.Name + " meta <file/dir1> <file/dir2> <file/dir3> ...",
			Description: "Default: metadata of current workdir",
			Category:    "Baidu Netdisk",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				var (
					ca = c.Args()
					as []string
				)
				if len(ca) == 0 {
					as = []string{""}
				} else {
					as = ca
				}

				pcscommand.RunGetMeta(as...)
				return nil
			},
		},
		{
			Name:      "rm",
			Usage:     "Delete file/directory",
			UsageText: app.Name + " rm <file/dir path1> <file/dir2> <file/dir3> ...",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunRemove(c.Args()...)
				return nil
			},
		},
		{
			Name:      "mkdir",
			Usage:     "Create directory",
			UsageText: app.Name + " mkdir <dir>",
			Category:  "Baidu Netdisk",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunMkdir(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:  "cp",
			Usage: "Copy file/directory",
			UsageText: `BaiduPCS-Go cp </> </>
	BaiduPCS-Go cp </1> </2> </3> ... <target dir>`,
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunCopy(c.Args()...)
				return nil
			},
		},
		{
			Name:  "mv",
			Usage: "Move/rename file/directory",
			UsageText: `:
	BaiduPCS-Go mv </1> </2> </3> ... <target dir>

	:
	BaiduPCS-Go mv </> </>`,
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunMove(c.Args()...)
				return nil
			},
		},
		{
			Name:      "download",
			Aliases:   []string{"d"},
			Usage:     "Download file/directory",
			UsageText: app.Name + " download <file/dir path1> <file/dir2> <file/dir3> ...",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				// saveTo
				var (
					saveTo string
				)
				if c.Bool("save") {
					saveTo = "."
				} else if c.String("saveto") != "" {
					saveTo = filepath.Clean(c.String("saveto"))
				}

				// downloadMode
				var (
					downloadMode pcsdownload.DownloadMode
				)
				switch c.String("mode") {
				case "pcs":
					downloadMode = pcsdownload.DownloadModePCS
				case "stream":
					downloadMode = pcsdownload.DownloadModeStreaming
				case "locate":
					downloadMode = pcsdownload.DownloadModeLocate
				default:
					fmt.Println("Failed to parse download mode")
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				do := &pcscommand.DownloadOptions{
					IsTest:               c.Bool("test"),
					IsPrintStatus:        c.Bool("status"),
					IsExecutedPermission: c.Bool("x"),
					IsOverwrite:          c.Bool("ow"),
					DownloadMode:         downloadMode,
					SaveTo:               saveTo,
					Parallel:             c.Int("p"),
					Load:                 c.Int("l"),
					MaxRetry:             c.Int("retry"),
					NoCheck:              c.Bool("nocheck"),
					LinkPrefer:           c.Int("dindex"),
					ModifyMTime:          c.Bool("mtime"),
					FullPath:             c.Bool("fullpath"),
				}

				pcscommand.RunDownload(c.Args(), do)

				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "test",
					Usage: "Test download, will not save files locally",
				},
				cli.BoolFlag{
					Name:  "ow",
					Usage: "overwrite existing files",
				},
				cli.BoolFlag{
					Name:  "status",
					Usage: "Print all worker statuses",
				},
				cli.BoolFlag{
					Name:  "save",
					Usage: "Save downloaded files to current workdir",
				},
				cli.StringFlag{
					Name:  "saveto",
					Usage: "Save downloaded files to specified directory",
				},
				cli.BoolFlag{
					Name:  "x",
					Usage: "Add executable permission (not effective on Windows)",
				},
				cli.StringFlag{
					Name:  "mode",
					Usage: "Download mode: pcs, stream, locate. Default is locate; see help above",
					Value: "locate",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "Set download thread count",
				},
				cli.IntFlag{
					Name:  "l",
					Usage: "Set number of files downloading simultaneously",
				},
				cli.IntFlag{
					Name:  "retry",
					Usage: "Max retry count for failed downloads",
					Value: pcsdownload.DefaultDownloadMaxRetry,
				},
				cli.BoolFlag{
					Name:  "nocheck",
					Usage: "Do not verify file after download",
				},
				cli.BoolFlag{
					Name:  "mtime",
					Usage: "Set local file mtime to server mtime",
				},
				cli.IntFlag{
					Name:  "dindex",
					Usage: "Choose which alternative link to use, default is first",
				},
				cli.BoolFlag{
					Name:  "fullpath",
					Usage: "Save using full cloud path locally",
				},
			},
		},
		{
			Name:      "upload",
			Aliases:   []string{"u"},
			Usage:     "Upload file/directory",
			UsageText: app.Name + " upload <local file/dir path1> <file/dir2> <file/dir3> ... <target dir>",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() < 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				subArgs := c.Args()
				pcscommand.RunUpload(subArgs[:c.NArg()-1], subArgs[c.NArg()-1], &pcscommand.UploadOptions{
					Parallel:      c.Int("p"),
					MaxRetry:      c.Int("retry"),
					Load:          c.Int("l"),
					NoRapidUpload: c.Bool("norapid"),
					Policy:        c.String("policy"),
				})
				return nil
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "p",
					Usage: "Set max upload threads per file",
				},
				cli.IntFlag{
					Name:  "retry",
					Usage: "Max retry count for failed uploads",
					Value: pcscommand.DefaultUploadMaxRetry,
				},
				cli.IntFlag{
					Name:  "l",
					Usage: "Set max number of files uploading simultaneously",
				},
				cli.BoolFlag{
					Name:  "norapid",
					Usage: "Skip rapid upload",
				},
				cli.StringFlag{
					Name:  "policy",
					Usage: fmt.Sprintf("Duplicate file policy (default: %s), %s, %s", baidupcs.SkipPolicy, baidupcs.OverWritePolicy, baidupcs.RsyncPolicy),
				},
			},
		},
		{
			Name:      "locate",
			Aliases:   []string{"lt"},
			Usage:     "Get direct download links",
			UsageText: app.Name + " locate <file1> <file2> ...",
			Description: fmt.Sprintf(`
	Get direct download links

	, "user is not authorized, hitcode:xxx",  User-Agent  %s:
	BaiduPCS-Go config set -user_agent "%s"
`, baidupcs.NetdiskUA, baidupcs.NetdiskUA),
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				opt := &pcscommand.LocateDownloadOption{
					FromPan: c.Bool("pan"),
				}

				pcscommand.RunLocateDownload(c.Args(), opt)
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "pan",
					Usage: "Get download links from Baidu Netdisk home API",
				},
			},
		},
		{
			Name:      "sumfile",
			Aliases:   []string{"sf"},
			Usage:     "Get local rapid-upload info (rapid-upload is currently unavailable)",
			UsageText: app.Name + " sumfile <local file path1> <local file path2> ...",
			Description: "See command usage and options for details.",
			Category: "Other",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				for k, filePath := range c.Args() {
					lp, err := checksum.GetFileSum(filePath, checksum.CHECKSUM_MD5|checksum.CHECKSUM_SLICE_MD5|checksum.CHECKSUM_CRC32)
					if err != nil {
						fmt.Printf("[%d] %s\n", k+1, err)
						continue
					}

					fmt.Printf("[%d] - [%s]:\n", k+1, filePath)

					strLength, strMd5, strSliceMd5, strCrc32 := strconv.FormatInt(lp.Length, 10), hex.EncodeToString(lp.MD5), hex.EncodeToString(lp.SliceMD5), strconv.FormatUint(uint64(lp.CRC32), 10)
					fileName := filepath.Base(filePath)
					regFileName := strings.Replace(fileName, " ", "_", -1)
					regFileName = strings.Replace(regFileName, "#", "_", -1)
					tb := pcstable.NewTable(os.Stdout)
					tb.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
					tb.AppendBulk([][]string{
						[]string{"File size", strLength},
						[]string{"md5", strMd5},
						[]string{"MD5 of first 256KB slice", strSliceMd5},
						[]string{"crc32", strCrc32},
						[]string{"Rapid upload command", app.Name + " rapidupload -length=" + strLength + " -md5=" + strMd5 + " -slicemd5=" + strSliceMd5 + " -crc32=" + strCrc32 + " " + fileName},
						[]string{"Generic rapid-upload link", strMd5 + "#" + strSliceMd5 + "#" + strLength + "#" + regFileName},
					})
					tb.Render()
					fmt.Printf("\n")
				}

				return nil
			},
		},
		{
			Name:      "transfer",
			Usage:     "Save shared file/directory to cloud",
			UsageText: app.Name + " transfer <share link> <extract code> (if any)",
			Category:  "Baidu Netdisk",
			Before:    reloadFn,
			Description: "See command usage and options for details.",
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 || c.NArg() > 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				opt := &baidupcs.TransferOption{
					Download: c.Bool("download"),
					Collect:  c.Bool("collect"),
					Rname:    c.Bool("rname"),
				}
				pcscommand.RunShareTransfer(c.Args(), opt)
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "download",
					Usage: "Download to local default dir after transfer",
				},
				cli.BoolFlag{
					Name:  "collect",
					Usage: "Collect multiple files into one folder when transferring",
				},
				cli.BoolFlag{
					Name:  "rname",
					Usage: "Randomly replace 4 chars in filename to improve rapid-transfer success",
				},
			},
		},
		{
			Name:      "share",
			Usage:     "Share file/directory",
			UsageText: app.Name + " share",
			Category:  "Baidu Netdisk",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:        "set",
					Aliases:     []string{"s"},
					Usage:       "Set share target",
					UsageText:   app.Name + " share set <file/dir1> <file/dir2> ...",
					Description: "See command usage and options for details.",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						opt := &baidupcs.ShareOption{
							Password:   c.String("p"),
							Period:     c.Int("period"),
							IsCombined: c.Bool("f"),
						}
						pcscommand.RunShareSet(c.Args(), opt)
						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "p",
							Usage: "Extract code",
							Value: "",
						},
						cli.IntFlag{
							Name:  "period",
							Usage: "Valid days, 0 means permanent",
							Value: 0,
						},
						cli.BoolFlag{
							Name:  "f",
							Usage: "Output full link format with password",
						},
					},
				},
				{
					Name:      "list",
					Aliases:   []string{"l"},
					Usage:     "List shared files/directories",
					UsageText: app.Name + " share list",
					Action: func(c *cli.Context) error {
						pcscommand.RunShareList(c.Int("page"))
						return nil
					},
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "page",
							Usage: "Page number for share list",
							Value: 1,
						},
					},
				},
				{
					Name:        "cancel",
					Aliases:     []string{"c"},
					Usage:       "Cancel sharing files/directories",
					UsageText:   app.Name + " share cancel <shareid_1> <shareid_2> ...",
					Description: "See command usage and options for details.",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						pcscommand.RunShareCancel(converter.SliceStringToInt64(c.Args()))
						return nil
					},
				},
			},
		},
		{
			Name:      "export",
			Aliases:   []string{"ep"},
			Usage:     "Export file/directory",
			UsageText: app.Name + " export <file/dir1> <file/dir2> ...",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				pcspaths := c.Args()
				if len(pcspaths) == 0 {
					pcspaths = []string{"."}
				}

				pcscommand.RunExport(pcspaths, &pcscommand.ExportOptions{
					RootPath:   c.String("root"),
					SavePath:   c.String("out"),
					MaxRetry:   c.Int("retry"),
					Recursive:  c.Bool("r"),
					LinkFormat: c.Bool("link"),
					StdOut:     c.Bool("stdout"),
				})
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "root",
					Usage: "Set root path for export (can be relative)",
				},
				cli.StringFlag{
					Name:  "out",
					Usage: "Output path for exported file info",
				},
				cli.IntFlag{
					Name:  "retry",
					Usage: "Retry count for export failures",
					Value: 3,
				},
				cli.BoolFlag{
					Name:  "r",
					Usage: "Export recursively",
				},
				cli.BoolFlag{
					Name:  "link",
					Usage: "Export in generic rapid-upload link format (path info will be lost)",
				},
				cli.BoolFlag{
					Name:  "stdout",
					Usage: "Do not write export to file, print to stdout",
				},
			},
		},
		{
			Name:    "offlinedl",
			Aliases: []string{"clouddl", "od"},
			Usage:   "Offline download",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "add",
					Aliases:   []string{"a"},
					Usage:     "Add offline download task",
					UsageText: app.Name + " offlinedl add -path=<offline save path> url1 url2 ...",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						pcscommand.RunCloudDlAddTask(c.Args(), c.String("path"))
						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "path",
							Usage: "Offline download save path, default is current workdir",
						},
					},
				},
				{
					Name:      "query",
					Aliases:   []string{"q"},
					Usage:     "Query offline download tasks by ID",
					UsageText: app.Name + " offlinedl query task_id1 task_id2 ...",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						taskIDs := converter.SliceStringToInt64(c.Args())

						if len(taskIDs) == 0 {
							fmt.Printf("No valid task ID found, task_id\n")
							return nil
						}

						pcscommand.RunCloudDlQueryTask(taskIDs)
						return nil
					},
				},
				{
					Name:      "list",
					Aliases:   []string{"ls", "l"},
					Usage:     "List offline download tasks",
					UsageText: app.Name + " offlinedl list",
					Action: func(c *cli.Context) error {
						pcscommand.RunCloudDlListTask()
						return nil
					},
				},
				{
					Name:      "cancel",
					Aliases:   []string{"c"},
					Usage:     "Cancel offline download tasks",
					UsageText: app.Name + " offlinedl cancel task_id1 task_id2 ...",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						taskIDs := converter.SliceStringToInt64(c.Args())

						if len(taskIDs) == 0 {
							fmt.Printf("No valid task ID found, task_id\n")
							return nil
						}

						pcscommand.RunCloudDlCancelTask(taskIDs)
						return nil
					},
				},
				{
					Name:      "delete",
					Aliases:   []string{"del", "d"},
					Usage:     "Delete offline download tasks",
					UsageText: app.Name + " offlinedl delete task_id1 task_id2 ...",
					Action: func(c *cli.Context) error {
						isClear := c.Bool("all")
						if c.NArg() < 1 && !isClear {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						// Offline download
						if isClear {
							pcscommand.RunCloudDlClearTask()
							return nil
						}

						// Offline download
						taskIDs := converter.SliceStringToInt64(c.Args())
						if len(taskIDs) == 0 {
							fmt.Printf("No valid task ID found, task_id\n")
							return nil
						}

						pcscommand.RunCloudDlDeleteTask(taskIDs)
						return nil
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "all",
							Usage: "Clear offline download task records; no second confirmation, use with caution!!!",
						},
					},
				},
			},
		},
		{
			Name:  "recycle",
			Usage: "Recycle bin",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NumFlags() <= 0 || c.NArg() <= 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
				}
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "list",
					Aliases:   []string{"ls", "l"},
					Usage:     baidupcs.OperationRecycleList,
					UsageText: app.Name + " recycle list",
					Action: func(c *cli.Context) error {
						pcscommand.RunRecycleList(c.Int("page"))
						return nil
					},
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "page",
							Usage: "Recycle bin list page",
							Value: 1,
						},
					},
				},
				{
					Name:        "restore",
					Aliases:     []string{"r"},
					Usage:       baidupcs.OperationRecycleRestore,
					UsageText:   app.Name + " recycle restore <fs_id 1> <fs_id 2> <fs_id 3> ...",
					Description: "See command usage and options for details.",
					Action: func(c *cli.Context) error {
						if c.NArg() <= 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						pcscommand.RunRecycleRestore(c.Args()...)
						return nil
					},
				},
				{
					Name:        "delete",
					Aliases:     []string{"d"},
					Usage:       baidupcs.OperationRecycleDelete + "/" + baidupcs.OperationRecycleClear,
					UsageText:   app.Name + " recycle delete [-all] <fs_id 1> <fs_id 2> <fs_id 3> ...",
					Description: "See command usage and options for details.",
					Action: func(c *cli.Context) error {
						if c.Bool("all") {
							// Recycle bin
							pcscommand.RunRecycleClear()
							return nil
						}

						if c.NArg() <= 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}
						pcscommand.RunRecycleDelete(c.Args()...)
						return nil
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "all",
							Usage: "Clear recycle bin; no second confirmation, use with caution!!!",
						},
					},
				},
			},
		},
		{
			Name:        "config",
			Usage:       "Show and modify config items",
			Description: "Show and modify config items",
			Category:    "Config",
			Before:      reloadFn,
			After:       saveFunc,
			Action: func(c *cli.Context) error {
				fmt.Printf("----\nRun %s config set to update config\n\nCurrent config:\n", app.Name)
				pcsconfig.Config.PrintTable()
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "set",
					Usage:     "Modify config items",
					UsageText: app.Name + " config set [arguments...]",
					Description: "See command usage and options for details.",
					Action: func(c *cli.Context) error {
						if c.NumFlags() <= 0 || c.NArg() > 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						if c.IsSet("appid") {
							pcsconfig.Config.SetAppID(c.Int("appid"))
						}
						if c.IsSet("enable_https") {
							pcsconfig.Config.SetEnableHTTPS(c.Bool("enable_https"))
						}
						if c.IsSet("ignore_illegal") {
							pcsconfig.Config.SetIgnoreIllegal(c.Bool("ignore_illegal"))
						}
						if c.IsSet("force_login_username") {
							pcsconfig.Config.SetForceLogin(c.String("force_login_username"))
						}
						if c.IsSet("user_agent") {
							pcsconfig.Config.SetUserAgent(c.String("user_agent"))
						}
						if c.IsSet("pcs_ua") {
							pcsconfig.Config.SetPCSUA(c.String("pcs_ua"))
						}
						if c.IsSet("pcs_addr") {
							match := pcsconfig.Config.SETPCSAddr(c.String("pcs_addr"))
							if !match {
								fmt.Println("Failed to set pcs_addr: invalid PCS server address")
								return nil
							}
						}
						if c.IsSet("fix_pcs_addr") {
							pcsconfig.Config.SetStaticPCSAddr(c.Bool("fix_pcs_addr"))
						}
						if c.IsSet("upload_policy") {
							pcsconfig.Config.SetUploadPolicy(c.String("upload_policy"))
						}
						if c.IsSet("pan_ua") {
							pcsconfig.Config.SetPanUA(c.String("pan_ua"))
						}
						if c.IsSet("cache_size") {
							err := pcsconfig.Config.SetCacheSizeByStr(c.String("cache_size"))
							if err != nil {
								fmt.Printf("Failed to set cache_size: %s\n", err)
								return nil
							}
						}
						if c.IsSet("max_parallel") {
							pcsconfig.Config.MaxParallel = c.Int("max_parallel")
						}
						if c.IsSet("max_upload_parallel") {
							pcsconfig.Config.MaxUploadParallel = c.Int("max_upload_parallel")
						}
						if c.IsSet("max_download_load") {
							pcsconfig.Config.MaxDownloadLoad = c.Int("max_download_load")
						}
						if c.IsSet("max_upload_load") {
							pcsconfig.Config.MaxUploadLoad = c.Int("max_upload_load")
						}
						if c.IsSet("max_download_rate") {
							err := pcsconfig.Config.SetMaxDownloadRateByStr(c.String("max_download_rate"))
							if err != nil {
								fmt.Printf("Failed to set max_download_rate: %s\n", err)
								return nil
							}
						}
						if c.IsSet("max_upload_rate") {
							err := pcsconfig.Config.SetMaxUploadRateByStr(c.String("max_upload_rate"))
							if err != nil {
								fmt.Printf("Failed to set max_upload_rate: %s\n", err)
								return nil
							}
						}
						if c.IsSet("savedir") {
							pcsconfig.Config.SaveDir = c.String("savedir")
						}
						if c.IsSet("proxy") {
							pcsconfig.Config.SetProxy(c.String("proxy"))
						}
						if c.IsSet("proxy_hostnames") {
							pcsconfig.Config.SetProxyHostnames(c.String("proxy_hostnames"))
						}
						if c.IsSet("local_addrs") {
							pcsconfig.Config.SetLocalAddrs(c.String("local_addrs"))
						}

						err := pcsconfig.Config.Save()
						if err != nil {
							fmt.Println(err)
							return err
						}

						pcsconfig.Config.PrintTable()
						fmt.Printf("\nConfig saved successfully!\n\n")

						return nil
					},
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "appid",
							Usage: "Baidu PCS AppID",
						},
						cli.StringFlag{
							Name:  "cache_size",
							Usage: "Download cache",
						},
						cli.IntFlag{
							Name:  "max_parallel",
							Usage: "Max total download concurrency",
						},
						cli.IntFlag{
							Name:  "max_upload_parallel",
							Usage: "Max upload concurrency per file",
						},
						cli.IntFlag{
							Name:  "max_download_load",
							Usage: "Max simultaneous downloading files",
						},
						cli.IntFlag{
							Name:  "max_upload_load",
							Usage: "Max simultaneous uploading files",
						},
						cli.StringFlag{
							Name:  "max_download_rate",
							Usage: "Limit max download speed, 0 for unlimited",
						},
						cli.StringFlag{
							Name:  "max_upload_rate",
							Usage: "Limit max upload speed, 0 for unlimited",
						},
						cli.StringFlag{
							Name:  "savedir",
							Usage: "Save directory for downloads",
						},
						cli.BoolFlag{
							Name:  "enable_https",
							Usage: "Enable HTTPS",
						},
						cli.BoolFlag{
							Name:  "ignore_illegal",
							Usage: "Ignore illegal characters in upload filename",
						},
						cli.StringFlag{
							Name:  "force_login_username",
							Usage: "Force login with specified username (for tieba API failure only)",
						},
						cli.BoolFlag{
							Name:  "no_check",
							Usage: "Disable MD5 verification after download",
						},
						cli.StringFlag{
							Name:  "upload_policy",
							Usage: "Policy for duplicate upload names",
						},
						cli.StringFlag{
							Name:  "user_agent",
							Usage: "User-Agent",
						},
						cli.StringFlag{
							Name:  "pcs_ua",
							Usage: "PCS User-Agent",
						},
						cli.StringFlag{
							Name:  "pcs_addr",
							Usage: "PCS server address",
						},
						cli.BoolFlag{
							Name:  "fix_pcs_addr",
							Usage: "Use static PCS server",
						},
						cli.StringFlag{
							Name:  "pan_ua",
							Usage: "Pan User-Agent",
						},
						cli.StringFlag{
							Name:  "proxy",
							Usage: "Set proxy, supports http/socks5",
						},
						cli.StringFlag{
							Name:  "proxy_hostnames",
							Usage: "Set proxied hostnames (comma-separated); empty means all",
						},
						cli.StringFlag{
							Name:  "local_addrs",
							Usage: "Set local interface addresses (comma-separated)",
						},
					},
				},
				{
					Name:        "reset",
					Usage:       "Restore default config",
					UsageText:   app.Name + " config reset",
					Description: "",
					Action: func(c *cli.Context) error {
						pcsconfig.Config.InitDefaultConfig()
						err := pcsconfig.Config.Save()
						if err != nil {
							fmt.Println(err)
							return err
						}
						pcsconfig.Config.PrintTable()
						fmt.Println("Default config restored")
						return nil
					},
				},
			},
		},
		{
			Name:      "match",
			Usage:     "Test wildcard",
			UsageText: app.Name + " match <wildcard expression>",
			Description: "See command usage and options for details.",
			Category: "Baidu Netdisk",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() != 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunTestShellPattern(c.Args()[0])
				return nil
			},
		},
		{
			Name:  "tool",
			Usage: "Toolbox",
			Action: func(c *cli.Context) error {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "showtime",
					Usage: "Show current time (Beijing time)",
					Action: func(c *cli.Context) error {
						fmt.Printf(pcstime.BeijingTimeOption("printLog"))
						return nil
					},
				},
				{
					Name:  "getip",
					Usage: "Get IP addresses",
					Action: func(c *cli.Context) error {
						fmt.Printf("LAN IP addresses: \n")
						for _, address := range pcsutil.ListAddresses() {
							fmt.Printf("%s\n", address)
						}
						fmt.Printf("\n")

						ipAddr, err := getip.IPInfoFromTechainBaiduByClient(pcsconfig.Config.HTTPClient())
						if err != nil {
							fmt.Printf("Failed to get public IP: %s\n", err)
							return nil
						}

						fmt.Printf("Public IP address: %s\n", ipAddr)
						return nil
					},
				},
				{
					Name:        "enc",
					Usage:       "Encrypt file",
					UsageText:   app.Name + " enc -method=<method> -key=<key> [files...]",
					Description: cryptoDescription,
					Action: func(c *cli.Context) error {
						if c.NArg() <= 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						for _, filePath := range c.Args() {
							encryptedFilePath, err := pcsutil.EncryptFile(c.String("method"), []byte(c.String("key")), filePath, !c.Bool("disable-gzip"))
							if err != nil {
								fmt.Printf("%s\n", err)
								continue
							}

							fmt.Printf("Encryption succeeded, %s -> %s\n", filePath, encryptedFilePath)
						}

						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "method",
							Usage: "Encryption method",
							Value: "aes-128-ctr",
						},
						cli.StringFlag{
							Name:  "key",
							Usage: "Encryption key",
							Value: app.Name,
						},
						cli.BoolFlag{
							Name:  "disable-gzip",
							Usage: "Disable GZIP",
						},
					},
				},
				{
					Name:        "dec",
					Usage:       "Decrypt file",
					UsageText:   app.Name + " dec -method=<method> -key=<key> [files...]",
					Description: cryptoDescription,
					Action: func(c *cli.Context) error {
						if c.NArg() <= 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						for _, filePath := range c.Args() {
							decryptedFilePath, err := pcsutil.DecryptFile(c.String("method"), []byte(c.String("key")), filePath, !c.Bool("disable-gzip"))
							if err != nil {
								fmt.Printf("%s\n", err)
								continue
							}

							fmt.Printf("Decryption succeeded, %s -> %s\n", filePath, decryptedFilePath)
						}

						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "method",
							Usage: "Encryption method",
							Value: "aes-128-ctr",
						},
						cli.StringFlag{
							Name:  "key",
							Usage: "Encryption key",
							Value: app.Name,
						},
						cli.BoolFlag{
							Name:  "disable-gzip",
							Usage: "Disable GZIP",
						},
					},
				},
			},
		},
		{
			Name:        "clear",
			Aliases:     []string{"cls"},
			Usage:       "Clear console",
			UsageText:   app.Name + " clear",
			Description: "Clear console screen",
			Category:    "Other",
			Action: func(c *cli.Context) error {
				pcsliner.ClearScreen()
				return nil
			},
		},
		{
			Name:    "quit",
			Aliases: []string{"exit"},
			Usage:   "Exit program",
			Action: func(c *cli.Context) error {
				return cli.NewExitError("", 0)
			},
			Hidden:   true,
			HideHelp: true,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}
