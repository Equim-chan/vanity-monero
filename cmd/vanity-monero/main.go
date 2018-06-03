package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	vanity "ekyu.moe/vanity-monero"
	"ekyu.moe/vanity-monero/mnemonic"
)

type matchMode uint8

const (
	mPrefix2 matchMode = iota
	mPrefix
	mRegex
)

func main() {
	var network vanity.Network
	fmt.Println("Select network:")
	fmt.Println("1) Monero main network")
	fmt.Println("2) Monero test network")
	switch promptNumber("Your choice:", 1, 2) {
	case 1:
		network = vanity.MoneroMainNetwork
	case 2:
		network = vanity.MoneroTestNetwork
	}
	fmt.Println()

	var dict *mnemonic.Dict
	fmt.Println("Select mnemonic seeds language:")
	fmt.Println("1) English")
	fmt.Println("2) Dutch")
	fmt.Println("3) Esperanto")
	fmt.Println("4) Spanish")
	fmt.Println("5) French")
	fmt.Println("6) German")
	fmt.Println("7) Italian")
	fmt.Println("8) Japanese")
	fmt.Println("9) Lojban")
	fmt.Println("10) Portuguese")
	fmt.Println("11) Russian")
	fmt.Println("12) Chinese (Simplified)")
	switch promptNumber("Your choice:", 1, 12) {
	case 1:
		dict = mnemonic.English
	case 2:
		dict = mnemonic.Dutch
	case 3:
		dict = mnemonic.Esperanto
	case 4:
		dict = mnemonic.Spanish
	case 5:
		dict = mnemonic.French
	case 6:
		dict = mnemonic.German
	case 7:
		dict = mnemonic.Italian
	case 8:
		dict = mnemonic.Japanese
	case 9:
		dict = mnemonic.Lojban
	case 10:
		dict = mnemonic.Portuguese
	case 11:
		dict = mnemonic.Russian
	case 12:
		dict = mnemonic.ChineseSimplified
	}
	fmt.Println()

	var mode matchMode
	var initIndex int
	fmt.Println("Select match mode:")
	fmt.Println(`1) Prefix from the 3rd character. (fast)
   For example, pattern "Ai" matches "42Aiabc..." and "48Aiabc...".
2) Prefix from the 1st character. (medium)
   For example, pattern "44Ai" matches "44Aiabc..." and "44Aidef...".
3) Regex. (slow)
   For example, pattern ".*A[0-9]{1,3}i.+" matches "44abcA233idef...".
   Note that in Regex mode there is no guarantee that there exists such address matching the pattern.`)
	switch promptNumber("Your choice:", 1, 3) {
	case 1:
		mode = mPrefix2
		initIndex = 2
	case 2:
		mode = mPrefix
		initIndex = 0
	case 3:
		mode = mRegex
	}
	fmt.Println()

PATTERN:
	var regex *regexp.Regexp
	var needOnlySpendKey bool
	pattern := prompt("Enter your prefix/regex:")
	switch mode {
	case mPrefix:
		if !vanity.IsValidPrefix(pattern, network, 0) {
			fmt.Println("invalid prefix")
			goto PATTERN
		}
		if len(pattern) < 2 {
			needOnlySpendKey = true
		} else {
			needOnlySpendKey = vanity.NeedOnlySpendKey(pattern[2:])
		}
	case mPrefix2:
		if !vanity.IsValidPrefix(pattern, network, 2) {
			fmt.Println("invalid prefix")
			goto PATTERN
		}
		needOnlySpendKey = vanity.NeedOnlySpendKey(pattern)
	case mRegex:
		var err error
		regex, err = regexp.Compile(pattern)
		if err != nil {
			fmt.Println("invalid regex:", err)
			goto PATTERN
		}
		needOnlySpendKey = false
	}

	caseSensitive := true
	if strings.ToLower(prompt("Case sensitive? [Y/n]")) == "n" {
		caseSensitive = false
		if mode == mRegex {
			regex = regexp.MustCompile("(?i)" + pattern)
		} else {
			pattern = strings.ToLower(pattern)
		}
	}

	n := promptNumber("Specify how many threads to run. 0 means all CPUs:", 0, 65535)
	runtime.GOMAXPROCS(n)
	threads := runtime.GOMAXPROCS(0)
	fmt.Println("=========================================")

	diff := uint64(0)
	switch mode {
	case mPrefix:
		diff = vanity.EstimatedDifficulty(pattern, caseSensitive, true)
	case mPrefix2:
		diff = vanity.EstimatedDifficulty(pattern, caseSensitive, false)
	}
	if diff == 0 {
		fmt.Println("Difficulty (est.): unknown")
	} else {
		fmt.Println("Difficulty (est.):", diff)
	}
	fmt.Println("Threads:", threads)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan *vanity.Key)
	ops := uint64(0)
	for i := 0; i < threads; i++ {
		// more code but less branches
		if mode == mRegex {
			go func() {
				seed := new([32]byte)
				key, addr := &vanity.Key{}, ""
				for ctx.Err() == nil {
					rand.Read(seed[:])

					key = vanity.KeyFromSeed(seed)
					addr = key.Address(network)

					if regex.MatchString(addr) {
						cancel()
						result <- key
						return
					}

					atomic.AddUint64(&ops, 1)
				}
			}()
		} else if needOnlySpendKey {
			if caseSensitive {
				go func() {
					seed := new([32]byte)
					key, addr := &vanity.Key{}, ""
					for ctx.Err() == nil {
						rand.Read(seed[:])

						key = vanity.HalfKeyFromSeed(seed)
						addr = key.HalfAddress(network)

						if strings.HasPrefix(addr[initIndex:], pattern) {
							cancel()
							result <- key
							return
						}

						atomic.AddUint64(&ops, 1)
					}
				}()
			} else {
				go func() {
					seed := new([32]byte)
					key, addr := &vanity.Key{}, ""
					for ctx.Err() == nil {
						rand.Read(seed[:])

						key = vanity.HalfKeyFromSeed(seed)
						addr = strings.ToLower(key.HalfAddress(network))

						if strings.HasPrefix(addr[initIndex:], pattern) {
							cancel()
							result <- key
							return
						}

						atomic.AddUint64(&ops, 1)
					}
				}()
			}
		} else {
			if caseSensitive {
				go func() {
					seed := new([32]byte)
					key, addr := &vanity.Key{}, ""
					for ctx.Err() == nil {
						rand.Read(seed[:])

						key = vanity.KeyFromSeed(seed)
						addr = key.Address(network)

						if strings.HasPrefix(addr[initIndex:], pattern) {
							cancel()
							result <- key
							return
						}

						atomic.AddUint64(&ops, 1)
					}
				}()
			} else {
				go func() {
					seed := new([32]byte)
					key, addr := &vanity.Key{}, ""
					for ctx.Err() == nil {
						rand.Read(seed[:])

						key = vanity.KeyFromSeed(seed)
						addr = strings.ToLower(key.Address(network))

						if strings.HasPrefix(addr[initIndex:], pattern) {
							cancel()
							result <- key
							return
						}

						atomic.AddUint64(&ops, 1)
					}
				}()
			}
		}
	}

	seconds := int64(0)
	keyrate := uint64(0)
	lastThree := uint64(0)
	padding := strings.Repeat(" ", 80)
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-t.C:
			seconds++
			last := atomic.LoadUint64(&ops)

			percentStr := "?"
			keyrateStr := "0"
			remainSecStr := "?"

			if diff > 0 {
				percent := float64(last) / float64(diff) * 100
				percentStr = strconv.FormatFloat(percent, 'f', 2, 64)
			}

			if keyrate != 0 {
				keyrateStr = strconv.FormatUint(keyrate, 10)
				if diff > last {
					remain := diff - last
					remainSecStr = (time.Duration(remain/keyrate) * time.Second).String()
				}
			}

			if seconds%3 == 0 {
				keyrate = (last - lastThree) / 3
				lastThree = last
			}

			stats := fmt.Sprintf("Key Rate: %s key/s || Total: %d (%s%%) || Time: %s / %s",
				keyrateStr,
				last,
				percentStr,
				time.Duration(seconds)*time.Second,
				remainSecStr,
			)
			fmt.Printf("%-.80s\r", stats+padding)

		case k := <-result:
			t.Stop()

			if needOnlySpendKey {
				k.HalfToFull()
			}
			words := dict.Encode(k.Seed())

			fmt.Println()
			fmt.Println("=========================================")
			fmt.Println("Address:")
			fmt.Println(k.Address(vanity.MoneroMainNetwork))
			fmt.Println()
			fmt.Println("Mnemonic Seed:")
			fmt.Println(strings.Join(words[:], " "))
			fmt.Println()
			fmt.Println("Private Spend Key:")
			fmt.Printf("%x\n", *k.SpendKey)
			fmt.Println()
			fmt.Println("Private View Key:")
			fmt.Printf("%x\n", *k.ViewKey)
			fmt.Println()
			fmt.Println()
			fmt.Println("HINT: You had better test the mnemonic seeds in Monero official wallet to check if they are legit. If the seeds work and you want to use the address, write the seeds down on real paper, and never disclose it!")
			fmt.Println()
			if runtime.GOOS == "windows" {
				fmt.Println("[Press Enter to exit]")
				stdin.Scan()
			}
			return
		}
	}
}
