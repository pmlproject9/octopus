package ipset

import (
	"fmt"
	"k8s.io/klog/v2"
	utilexec "k8s.io/utils/exec"
	"regexp"
	"strconv"
	"strings"
)

var IPSetCmd = "ipset"

var EntryMemberPattern = "(?m)^(.*\n)*Members:\n"

type Executor struct {
	exec utilexec.Interface
}

func NewExecutor(exec utilexec.Interface) *Executor {
	return &Executor{
		exec: exec,
	}
}

func (e *Executor) run(args []string, errFormat string) (string, error) {
	out, err := e.exec.Command(IPSetCmd, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s:%w (%s)", errFormat, err, out)
	}
	return string(out), nil
}

func (e *Executor) Create(ipset *IPSet) error {
	args := []string{"create", ipset.Name, ipset.HashType, "family", ipset.HashFamily,
		"hashsize", strconv.Itoa(ipset.HashSize), "-exist"}
	_, err := e.run(args, fmt.Sprintf("error creating ipset %s", ipset.Name))
	return err
}

func (e *Executor) Destroy(name string) error {
	args := []string{"destroy", name}
	_, err := e.run(args, fmt.Sprintf("error destroy ipset:%s", name))
	return err
}

func (e *Executor) DestroyIfExist(name string) error {
	ipsets, err := e.ListIPSets()
	if err != nil {
		return err
	}
	existed := false
	for _, ipset := range ipsets {
		if ipset == name {
			existed = true
		}
	}
	if existed {
		if err := e.Destroy(name); err != nil {
			return err
		}
	}
	return nil
}

func (e *Executor) Add(name string, entry string, timeout int) error {
	args := []string{"add", name, entry, "-exist"}
	_, err := e.run(args, fmt.Sprintf("error add entry into ipset:%s", name))
	return err
}

func (e *Executor) Del(name string, entry string) error {
	args := []string{"del", name, entry}
	_, err := e.run(args, fmt.Sprintf("error delete entry from ipset:%s", name))
	return err
}

func (e *Executor) Flush(name string) error {
	args := []string{"flush", name}
	_, err := e.run(args, fmt.Sprintf("error flush ipset:%s", name))
	return err
}

func (e *Executor) ListIPSets() ([]string, error) {
	args := []string{"list", "-n"}
	out, err := e.run(args, fmt.Sprintf("error list ipset"))
	if err != nil {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

func (e *Executor) ListEntries(name string) ([]string, error) {
	if name == "" {
		return nil, fmt.Errorf("ipset name can't be empty")
	}
	args := []string{"list", name}
	out, err := e.run(args, fmt.Sprintf("error list ipset: %s", name))
	if err != nil {
		return nil, err
	}
	memMatcher := regexp.MustCompile(EntryMemberPattern)
	memMatcher.ReplaceAllString(out, "")
	results := strings.Split(out, "\n")
	entries := make([]string, 0)
	for _, res := range results {
		if len(res) > 0 {
			entries = append(entries, res)
		}
	}
	return entries, nil
}

func (e *Executor) ReFlush(s *IPSet, entries []string) error {
	tempName := s.Name + "-t"
	tmepIPset := New(tempName, s.HashType, nil)
	err := e.Create(tmepIPset)
	if err != nil {
		klog.Errorf("error to create ipset %s (%v)", tmepIPset.Name, err)
		return err
	}
	err = e.Create(s)
	if err != nil {
		klog.Errorf("error to create ipset %s (%v)", s.Name, err)
		return err
	}
	for _, entry := range entries {
		err = e.Add(tmepIPset.Name, entry, 0)
		if err != nil {
			klog.Errorf("error addding entry %s to set: %s (%v)", entry, tmepIPset.Name, err)
			return err
		}
	}
	err = e.Swap(tmepIPset.Name, s.Name)
	if err != nil {
		return err
	}
	err = e.Destroy(tmepIPset.Name)
	if err != nil {
		return err
	}
	return nil
}

func (e *Executor) Swap(from string, to string) error {
	args := []string{"swap", from, to}
	_, err := e.run(args, fmt.Sprintf("error swap %s %s", from, to))
	return err
}
