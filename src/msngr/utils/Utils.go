package utils
import (

	"fmt"
	"math/rand"
	"reflect"
	"time"
	"regexp"
	"strings"

	"os"
	"log"
	"net/http"
	"io/ioutil"
	"path"
)


func GenId() string {
	t := time.Now().UnixNano()
	s := rand.NewSource(t)
	r := rand.New(s)
	return fmt.Sprintf("%d", r.Int63())
}

func FoundFile(fname string) *string {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}
	for {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil
		}
		for _, f := range files {
			if fname == f.Name() {
				result := path.Join(dir, fname)
				return &result
			}
		}
		dir = path.Dir(dir)

	}
	return nil
}

func ToMap(in interface{}, tag string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts structs; got %T", v)
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv := fi.Tag.Get(tag); tagv != "" {
			out[tagv] = v.Field(i).Interface()
		}
	}
	return out, nil
}

func FirstOf(data ...interface{}) interface{} {
	for _, data_el := range data {
		if data_el != "" {
			return data_el
		}
	}
	return ""
}

func In(p int, a []int) bool {
	for _, v := range a {
		if p == v {
			return true
		}
	}
	return false
}

func InS(p string, a []string) bool {
	for _, v := range a {
		if p == v {
			return true
		}
	}
	return false
}

func IntersectionS(a1, a2 []string) bool {
	for _, v1 := range a1 {
		for _, v2 := range a2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}

func Contains(container string, elements []string) bool {
	container_elements := regexp.MustCompile("[a-zA-Zа-яА-Я]+").FindAllString(container, -1)
	ce_map := make(map[string]bool)
	for _, ce_element := range container_elements {
		ce_map[strings.ToLower(ce_element)] = true
	}
	result := true
	for _, element := range elements {
		_, ok := ce_map[strings.ToLower(element)]
		result = result && ok
	}
	return result
}

func SaveToFile(what, fn string) {
	f, err := os.OpenFile(fn, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0600)
	if err != nil {
		log.Printf("ERROR when save to file in open file %v [%v]", fn, err)
	}

	defer f.Close()

	if _, err = f.WriteString(what); err != nil {
		log.Printf("ERROR when save to file in write to file %v [%v]", fn, err)
	}
}

func GET(url string, params *map[string]string) (*[]byte, error) {
	//	log.Printf("GET > [%+v] |%+v|", url, params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR IN GET FORM REQUEST! [%v]\n", url, err)
		return nil, err
	}

	if params != nil {
		values := req.URL.Query()
		for k, v := range *params {
			values.Add(k, v)
		}
		req.URL.RawQuery = values.Encode()
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if res == nil || err != nil {
		log.Println("ERROR IN GET DO REQUEST!\nRESPONSE: ", res, "\nERROR: ", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	//	log.Printf("GET < \n%v\n", string(body), )
	return &body, err
}
