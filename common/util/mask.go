package util

import "strings"

func MaskPhone(phone string) string {
	if n := len(phone); n >= 8 {
		return phone[:n-8] + "****" + phone[n-4:]
	}
	return phone
}

func MaskEmail(address string) string {
	strings.LastIndex(address, "@")
	id := address[0:strings.LastIndex(address, "@")]
	domain := address[strings.LastIndex(address, "@"):]
	if len(id) <= 1 {
		return address
	}
	switch len(id) {
	case 2:
		id = id[0:1] + "*"
	case 3:
		id = id[0:1] + "*" + id[2:]
	case 4:
		id = id[0:1] + "**" + id[3:]
	default:
		masks := strings.Repeat("*", len(id)-4)
		id = id[0:2] + masks + id[len(id)-2:]
	}
	return id + domain
}

func MaskRealName(realName string) string {
	runeRealName := []rune(realName)
	if n := len(runeRealName); n >= 2 {
		if n == 2 {
			return string(append(runeRealName[0:1], rune('*')))
		} else {
			count := n - 2
			newRealName := runeRealName[0:1]
			for temp := 1; temp < count; temp++ {
				newRealName = append(newRealName, rune('*'))
			}
			return string(append(newRealName, runeRealName[n-1]))
		}
	}
	return realName
}

func MaskLoginName(loginName string) string {
	if strings.LastIndex(loginName, "@") != -1 {
		return MaskEmail(loginName)
	}
	return MaskPhone(loginName)
}
