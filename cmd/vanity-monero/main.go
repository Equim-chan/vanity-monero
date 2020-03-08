package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	vanity "ekyu.moe/vanity-monero"
	"ekyu.moe/vanity-monero/mnemonic"
)

type (
	workMode  uint8
	matchMode uint8
)

const (
	wmStandard workMode = iota
	wmSplitKey

	mmPrefix2 matchMode = iota
	mmPrefix
	mmRegex
)

var stdin = bufio.NewScanner(os.Stdin)

func main() {
	var wMode workMode
	var partnerSpendPub, partnerViewPub *[32]byte // For split-key construct
	var myKey, partnerKey *vanity.Key             // For split-key restore
	fmt.Println("Select work mode:")
	fmt.Println("1) Standard mode")
	fmt.Println("2) Split-key mode")
	fmt.Println("3) About")
	switch promptNumber("Your choice:", 1, 3) {
	case 1:
		wMode = wmStandard

	case 2:
		wMode = wmSplitKey
		fmt.Println("Select your circumstance:")
		fmt.Println("1) I want someone else (can be untrusted) to generate a vanity address for me")
		fmt.Println("2) I want to help someone else to generate a vanity address")
		fmt.Println("3) Someone just generated a vanity address for me")
		switch promptNumber("Your choice:", 1, 3) {
		case 1:
			seed := new([32]byte)
			rand.Read(seed[:])
			key := vanity.KeyFromSeed(seed)
			pubSpend, pubView := key.PublicSpendKey(), key.PublicViewKey()
			fmt.Println()
			fmt.Println("=========================================")
			fmt.Println("Belows are your partial private key. Save and keep it secret and NEVER DISCLOSE!")
			fmt.Println("Partial Private Spend Key:")
			fmt.Printf("%x\n", *key.SpendKey)
			fmt.Println()
			fmt.Println("Partial Private View Key:")
			fmt.Printf("%x\n", *key.ViewKey)
			fmt.Println()
			fmt.Println("=========================================")
			fmt.Println("Belows are your partial public key. Hand it to your partner and also tell him the pattern you want.")
			fmt.Println("Partial Public Spend Key:")
			fmt.Printf("%x\n", *pubSpend)
			fmt.Println()
			fmt.Println("Partial Public View Key:")
			fmt.Printf("%x\n", *pubView)
			fmt.Println()
			exit()

		case 2:
			partnerSpendPub = prompt256Key("Enter your partner's public spend key:")
			partnerViewPub = prompt256Key("Enter your partner's public view key:")

		case 3:
			myKey = &vanity.Key{SpendKey: new([32]byte), ViewKey: new([32]byte)}
			myKey.SpendKey = prompt256Key("Enter your partial private spend key:")
			myKey.ViewKey = prompt256Key("Enter your partial private view key:")

			partnerKey = &vanity.Key{SpendKey: new([32]byte), ViewKey: new([32]byte)}
			partnerKey.SpendKey = prompt256Key("Enter your partner's partial private spend key:")
			partnerKey.ViewKey = prompt256Key("Enter your partner's partial private view key:")
		}

	case 3:
		fmt.Println()
		fmt.Println("=========================================")
		fmt.Println("Name: vanity-monero")
		fmt.Println("Description: Generate vanity address for CryptoNote currency (Monero etc.).")
		fmt.Println("Repo: https://github.com/Equim-chan/vanity-monero")
		fmt.Println("License: MIT")
		fmt.Println()
		fmt.Println("If you love this idea, maybe you can consider buying me a cup coffee at")
		fmt.Println("4777777jHFbZB4gyqrB1JHDtrGFusyj4b3M2nScYDPKEM133ng2QDrK9ycqizXS2XofADw5do5rU19LQmpTGCfeQTerm1Ti")
		exit()
	}
	fmt.Println()

	var network vanity.Network
	network = vanity.AvrioMainNetwork
	fmt.Println()

	var dict *mnemonic.Dict
	if partnerViewPub == nil {
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
	}

	if wMode == wmSplitKey && partnerKey != nil {
		// Restore the keys
		finalKey := myKey.Add(partnerKey)
		words := dict.Encode(finalKey.Seed())
		fmt.Println()
		fmt.Println("=========================================")
		fmt.Println("Address:")
		fmt.Println(finalKey.Address(network))
		fmt.Println()
		fmt.Println("Mnemonic Seed:")
		fmt.Println(strings.Join(words[:], " "))
		fmt.Println()
		fmt.Println("Private Spend Key:")
		fmt.Printf("%x\n", *finalKey.SpendKey)
		fmt.Println()
		fmt.Println("Private View Key:")
		fmt.Printf("%x\n", *finalKey.ViewKey)
		fmt.Println()
		fmt.Println()
		fmt.Println("HINT: You had better test the mnemonic seeds in Avrio's official wallet to check if they are correct. If the seeds work and you want to use the address, write the seeds down on real paper, and never disclose it!")
		exit()
	}

	var mMode matchMode
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
		mMode = mmPrefix2
		initIndex = 2
	case 2:
		mMode = mmPrefix
		initIndex = 0
	case 3:
		mMode = mmRegex
	}
	fmt.Println()

PATTERN:
	var regex *regexp.Regexp
	var needOnlySpendKey bool
	pattern := prompt("Enter your prefix/regex, which must be in ASCII and not include 'I', 'O', 'l':")
	switch mMode {
	case mmPrefix:
		if !vanity.IsValidPrefix(pattern, network, 0) {
			fmt.Println("invalid prefix")
			goto PATTERN
		}
		if len(pattern) < 2 {
			needOnlySpendKey = true
		} else {
			needOnlySpendKey = vanity.NeedOnlySpendKey(pattern[2:])
		}
	case mmPrefix2:
		if !vanity.IsValidPrefix(pattern, network, 2) {
			fmt.Println("invalid prefix")
			goto PATTERN
		}
		needOnlySpendKey = vanity.NeedOnlySpendKey(pattern)
	case mmRegex:
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
		if mMode == mmRegex {
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
	switch mMode {
	case mmPrefix:
		diff = vanity.EstimatedDifficulty(pattern, caseSensitive, true)
	case mmPrefix2:
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
	defer cancel()
	result := make(chan *vanity.Key)
	ops := uint64(0)
	for i := 0; i < threads; i++ {
		// more code but less branches
		if wMode == wmStandard {
			if mMode == mmRegex {
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
			} else { // Need full key
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
		} else { // Split-key mode
			if mMode == mmRegex {
				go func() {
					seed := new([32]byte)
					key, addr := &vanity.Key{}, ""
					for ctx.Err() == nil {
						rand.Read(seed[:])

						key = vanity.KeyFromSeed(seed)
						addr = key.AddressWithAdditionalPublicKey(network, partnerSpendPub, partnerViewPub)

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
							addr = key.HalfAddressWithAdditionalPublicKey(network, partnerSpendPub)

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
							addr = strings.ToLower(key.HalfAddressWithAdditionalPublicKey(network, partnerSpendPub))

							if strings.HasPrefix(addr[initIndex:], pattern) {
								cancel()
								result <- key
								return
							}

							atomic.AddUint64(&ops, 1)
						}
					}()
				}
			} else { // Need full key
				if caseSensitive {
					go func() {
						seed := new([32]byte)
						key, addr := &vanity.Key{}, ""
						for ctx.Err() == nil {
							rand.Read(seed[:])

							key = vanity.KeyFromSeed(seed)
							addr = key.AddressWithAdditionalPublicKey(network, partnerSpendPub, partnerViewPub)

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
							addr = strings.ToLower(key.AddressWithAdditionalPublicKey(network, partnerSpendPub, partnerViewPub))

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

			fmt.Println()
			fmt.Println("=========================================")
			if wMode == wmStandard {
				words := dict.Encode(k.Seed())
				fmt.Println("Address:")
				fmt.Println(k.Address(network))
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
			} else { // Split-key mode
				fmt.Println("Final Address:")
				fmt.Println(k.AddressWithAdditionalPublicKey(network, partnerSpendPub, partnerViewPub))
				fmt.Println()
				fmt.Println("Partial Private Spend Key:")
				fmt.Printf("%x\n", *k.SpendKey)
				fmt.Println()
				fmt.Println("Partial Private View Key:")
				fmt.Printf("%x\n", *k.ViewKey)
				fmt.Println()
				fmt.Println()
				fmt.Println("Give your partner the private keys shown above.")
			}

			exit()
		}
	}
}

func prompt(question string) string {
	for {
		fmt.Print(question + " ")
		stdin.Scan()
		ans := strings.TrimSpace(stdin.Text())
		if ans != "" {
			return ans
		}
		fmt.Println("can't be empty")
	}
}

func promptComfirm(question string) bool {
	return prompt(question) == "y"
}

func promptNumber(question string, min, max int) int {
	for {
		n, err := strconv.Atoi(prompt(question))
		switch {
		case err != nil:
			fmt.Println("invalid number")
		case n < min || n > max:
			fmt.Println("invalid range")
		default:
			return n
		}
	}
}

func prompt256Key(question string) *[32]byte {
	for {
		keyHex := prompt(question)
		if len(keyHex) != 64 {
			fmt.Println("Wrong key size, should be exactly 64 characters")
			continue
		}
		raw, err := hex.DecodeString(keyHex)
		if err != nil {
			fmt.Println(err)
			continue
		}

		ret := new([32]byte)
		copy(ret[:], raw)

		return ret
	}
}

func exit() {
	fmt.Println()
	if runtime.GOOS == "windows" {
		fmt.Println("[Press Enter to exit]")
		stdin.Scan()
	}
	os.Exit(0)
}
