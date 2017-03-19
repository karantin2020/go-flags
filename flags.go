package flags

import (
	"regexp"
	"reflect"
	"fmt"
	"strings"
	"os"
	"log"
	"strconv"
)

var (
	long = regexp.MustCompile("^--([A-Za-z]+[0-9A-Za-z]*)[=|:| +]([0-9A-Za-z./]+)$")
	longBool = regexp.MustCompile("^--([A-Za-z]+[0-9A-Za-z]*)$")
	short = regexp.MustCompile("^-([A-Za-z])[=|:| +]([0-9A-Za-z./]+)$")
	shortBool = regexp.MustCompile("^-([A-Za-z])$")

	shortPack = regexp.MustCompile("^-([A-Za-z]{2,})$")

	longCheck = regexp.MustCompile("^([A-Za-z]+[0-9A-Za-z]*)$")
	shortCheck = regexp.MustCompile("^([A-Za-z])$")
)

var (
	m Flags
)

type Flags struct {
	m map[string]*Item
	flags map[string]Flag
}

type Flag struct {
	Short string
	Long string
	Dst interface{}
	Req bool
	Def interface{}
	Do func()
}

type Item struct {
	Idx []int
	Vals []string
}

func parseFlags() map[string]*Item {
	m := make(map[string]*Item)
	next := false
	counter := 0
	for i, arg := range os.Args {
		if next {
			next = false
			continue
		}
		// fmt.Println(arg)
		dst := arg
		ws := false
		if !strings.Contains(arg, ":") && !strings.Contains(arg, "=") {
			if (i+1) < len(os.Args) {
				dst = arg + " " + os.Args[i+1]
				ws = true
			}
		}
		if lst := long.FindStringSubmatch(dst); len(lst) > 0 {
			if _, pres := m[lst[1]]; !pres {
				m[lst[1]] = &Item{}
			}
			v := m[lst[1]]
			v.Vals = append(v.Vals, lst[2])
			v.Idx = append(v.Idx, counter)
			counter++
			if ws {
				next = true
			}
			continue
		}
		if lst := short.FindStringSubmatch(dst); len(lst) > 0 {
			if _, pres := m[lst[1]]; !pres {
				m[lst[1]] = &Item{}
			}
			v := m[lst[1]]
			v.Vals = append(v.Vals, lst[2])
			v.Idx = append(v.Idx, counter)
			counter++
			if ws {
				next = true
			}
			continue
		}
		if lst := longBool.FindStringSubmatch(arg); len(lst) > 0 {
			if _, pres := m[lst[1]]; !pres {
				m[lst[1]] = &Item{}
			}
			v := m[lst[1]]
			v.Vals = append(v.Vals, "true")
			v.Idx = append(v.Idx, counter)
			counter++
			continue
		}
		if lst := shortBool.FindStringSubmatch(arg); len(lst) > 0 {
			if _, pres := m[lst[1]]; !pres {
				m[lst[1]] = &Item{}
			}
			v := m[lst[1]]
			v.Vals = append(v.Vals, "true")
			v.Idx = append(v.Idx, counter)
			counter++
			continue
		}

		if lst := shortPack.FindStringSubmatch(arg); len(lst) > 0 {
			for _, sn := range strings.Split(lst[1], "") {
				if _, pres := m[sn]; !pres {
					m[sn] = &Item{}
				}
				v := m[sn]
				v.Vals = append(v.Vals, "true")
				v.Idx = append(v.Idx, counter)
				counter++
			}
			continue
		}
	}
	return m
}

func Set(in Flag) {
	found := false
	var res string
	if in.Short == "" && in.Long == "" {
		log.Fatal("Empty flag passed")
	}
	// check for struct fields correctness
	if in.Long != "" && !longCheck.MatchString(in.Long) {
		log.Fatal("Incorrect flag name: " + in.Long)
	}
	if in.Short != "" && !shortCheck.MatchString(in.Short) {
		log.Fatal("Incorrect flag name: " + in.Short)
	}
	
	v := reflect.ValueOf(in.Dst)
	outt := v.Type()
	outk := outt.Kind()
	if reflect.TypeOf(in.Dst).Kind() != reflect.Ptr {
		log.Fatal("Unmarshal needs a pointer")
	}

	for {
		if outk == reflect.Ptr && v.IsNil() {
			if v.CanAddr() {
				v.Set(reflect.New(outt.Elem()))
			} else {
				log.Fatal("Unmarshal needs addressable pointer to a struct")
			}
		}
		if outk == reflect.Ptr {
			v = v.Elem()
			outt = v.Type()
			outk = v.Kind()
			continue
		}
		break
	}

	var sok, sb, lok, lb bool
	_, sok = m.m[in.Short]
	sb = !sok && outk == reflect.Bool
	_, lok = m.m[in.Long]
	lb = !lok && outk == reflect.Bool

	switch "" {
	case in.Short:
		if lok {
			res = in.Long
			m.flags[in.Long] = in
			found = true
		} else if lb {
			v.SetBool(false)
			return
		}
	case in.Long:
		if sok {
			res = in.Short
			m.flags[in.Short] = in
			found = true
		} else if sb {
			v.SetBool(false)
			return
		}
	default:
		switch outk {
		case reflect.Slice:
			if lok && sok {
				lv := m.m[in.Long]
				sv := m.m[in.Short]
				sv.Vals = append(sv.Vals, lv.Vals...)
				sv.Idx = append(sv.Idx, lv.Idx...)
				res = in.Short
				m.flags[in.Short] = in
				found = true
			} else if lok {
				res = in.Long
				m.flags[in.Long] = in
				found = true
			} else if sok {
				res = in.Short
				m.flags[in.Short] = in
				found = true
			}
		default:
			if lok && sok {
				res = findLast(in.Short, in.Long)
				m.flags[res] = in
				found = true
			} else if lok {
				res = in.Long
				m.flags[in.Long] = in
				found = true
			} else if sok {
				res = in.Short
				m.flags[in.Short] = in
				found = true
			} else if lb || sb {
				v.SetBool(false)
				return
			}
		}
	}

	if !found && in.Req {
		help("Didn't find flag \"" + in.Short + "/" + in.Long + "\" in arguments")
	}

	if !found {
		// fmt.Println("Not found", in.Short, in.Long)
		v.Set(reflect.ValueOf(in.Def))
		return
	}

	vm := m.m[res]
	vval := vm.Vals[len(vm.Vals)-1]
	switch v.Kind() {
	// Need to implement slice kind
	case reflect.Bool:
		if x, err := strconv.ParseBool(vval); err == nil {
			// fmt.Println("Set:", res, x)
			v.SetBool(x)
		} else {
			log.Fatal("Type error: " + res + ":" + vval)
		}
		
	case reflect.Float64:
		if x, err := strconv.ParseFloat(vval, 64); err == nil {
			v.SetFloat(x)
		} else {
			log.Fatal("Type error: " + res + ":" + vval)
		}
	case reflect.Int:
		fallthrough
	case reflect.Int64:
		if x, err := strconv.ParseInt(vval, 10, 64); err == nil {
			v.SetInt(x)
		} else {
			log.Fatal("Type error: " + res + ":" + vval)
		}
	case reflect.String:
		v.SetString(vval)
	case reflect.Uint64:
		if x, err := strconv.ParseUint(vval, 10, 64); err == nil {
			v.SetUint(x)
		} else {
			log.Fatal("Type error: " + res + ":" + vval)
		}
	default:
		log.Fatal("Type of flag value is not supported")
	}
}

func help(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func findLast(short, long string) string {
	s := m.m[short]
	l := m.m[long]
	if s.Idx[len(s.Idx)-1] > l.Idx[len(l.Idx)-1] {
		return short
	} else {
		return long
	}
}

func Get(name string) *Item {
	return m.m[name]
}

func GetNames() (res []string) {
	for k, _ := range m.m {
		res = append(res, k)
	}
	return
}

func init() {
	m.flags = make(map[string]Flag)
	m.m = parseFlags()
	// fmt.Printf("%#v\n", m.m)
}

/*
func main() {
	// fmt.Printf("%#v\n", os.Args)
	// fmt.Printf("%#v\n", m.m)
	var q int
	Add(Flag{
		Short: "f",
		Long: "",
		Dst: &q,
		Do: nil,
	})
	fmt.Println(q)
	var b bool
	Add(Flag{
		Short: "b",
		Long: "boo",
		Dst: &b,
		Do: nil,
	})
	fmt.Println(b)
}
*/