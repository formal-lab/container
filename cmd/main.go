package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"time"
)

const cgroupV2MemoryHierarchy = "/sys/fs/cgroup/user.slice"
const linuxSelfProc = "/proc/self/exe"

func main() {
	if os.Args[0] == linuxSelfProc {
		// make sure the sub process is created after the pid of parent process
		// is added into the cgroup.procs
		time.Sleep(1 * time.Second)

		// container process
		fmt.Printf("current pid: %d", syscall.Getpid())
		fmt.Println()
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// cmd := exec.Command("sh")
	cmd := exec.Command(linuxSelfProc)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWNET,
	}
	// cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(1), Gid: uint32(1)}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// if err := cmd.Run(); err != nil {
	// 	log.Fatal(err)
	// }
	// os.Exit(-1)
	if err := cmd.Start(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	} else {
		// get the pid of the forked process mapped in the outer namespace
		fmt.Printf("outer pid: %v", cmd.Process.Pid)
		fmt.Println()

		const memSubsystem = "memorylimit"
		// create a cgroup for the process on the default Hierarchy, which is
		// created by the OS
		os.Mkdir(path.Join(cgroupV2MemoryHierarchy, memSubsystem), 0755)
		// join the container to the cgroup
		os.WriteFile(path.Join(cgroupV2MemoryHierarchy, memSubsystem, "cgroup.procs"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
		// limit the cgroup process memory usage
		os.WriteFile((path.Join(cgroupV2MemoryHierarchy, memSubsystem, "memory.max")), []byte("100m"), 0644)
	}
	cmd.Process.Wait()
}
