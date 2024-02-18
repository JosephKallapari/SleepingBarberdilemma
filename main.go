package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type BarberShop struct {
	NoOfBarbers            int
	NoOfChairs             int
	hairCutDuraion         time.Duration
	ShopOpen               bool
	BarberhairCutCompleted chan bool
	NoOfCustomers          chan string
}

func main() {
	args := os.Args
	usage := fmt.Sprintf(""+
		"Usage: \n"+
		"\t%s <waiting room capacity> <haircut time in milliseconds> <average arrival rate in milliseconds> <shop open time in seconds>", args[0])
	if len(args) != 5 {
		fmt.Println(usage)
	} else {
		seatingCapacity, err1 := strconv.Atoi(args[1])
		timePerHairCut, err2 := strconv.Atoi(args[2])
		arrivalRate, err3 := strconv.Atoi(args[3])
		openTime, err4 := strconv.Atoi(args[4])
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			fmt.Println(err1, err2, err3, err4, usage)
		} else {
			runOperations(
				seatingCapacity,
				time.Millisecond*time.Duration(timePerHairCut),
				arrivalRate,
				time.Second*time.Duration(openTime),
			)
		}
	}
}

func runOperations(
	seatingCapacity int,
	timePerHairCut time.Duration,
	arrivalRate int,
	openTime time.Duration,
) {
	shop := NewBarberShop(seatingCapacity, timePerHairCut)
	shop.AddBarber("Barber")
	i := 1

	shopClosing := make(chan bool)
	closed := make(chan bool)

	go func() {
		<-time.After(openTime)
		shopClosing <- true
		shop.CloseShop()
		closed <- true
	}()

	go func() {
		for {
			// Get a random number with average arrival rate as specified
			randomMilliseconds := rand.Int() % (2 * arrivalRate)
			select {
			case <-shopClosing:
				return
			case <-time.After(time.Millisecond * time.Duration(randomMilliseconds)):
				shop.AddCustomers(strconv.Itoa(i))
				i++
			}
		}
	}()

	<-closed
}

func NewBarberShop(NoOfChairsInput int, hairCutDurationInput time.Duration) BarberShop {
	shop := BarberShop{NoOfBarbers: 0, NoOfChairs: NoOfChairsInput, hairCutDuraion: hairCutDurationInput, ShopOpen: true}
	shop.NoOfCustomers = make(chan string, shop.NoOfChairs)
	shop.BarberhairCutCompleted = make(chan bool)
	return shop
}

func (shop *BarberShop) CloseShop() {
	shop.ShopOpen = false
	fmt.Println("Closing Shop")
	close(shop.NoOfCustomers)
	for i := 0; i < shop.NoOfBarbers; i++ {
		<-shop.BarberhairCutCompleted
	}
	close(shop.BarberhairCutCompleted)
	fmt.Println("Shop Closed")
}

func (shop *BarberShop) AddCustomers(customer string) {
	if shop.ShopOpen {
		select {
		case shop.NoOfCustomers <- customer:
			fmt.Println("Add Customer %s", customer)
		default:
			fmt.Println("Shop is full please come back later")
		}
	} else {
		fmt.Println("Shop Closed cannot accept customers added")
	}
}

func (shop *BarberShop) AddBarber(barber string) {
	shop.NoOfBarbers++
	go func() {
		for {
			if len(shop.NoOfCustomers) == 0 {
				fmt.Printf("%s is sleeping: Zzzzzzzzzzzzz ...\n", barber)
			}
			client, shopOpen := <-shop.NoOfCustomers
			if shopOpen {
				shop.cutHair(barber, client)
			} else {
				shop.sendBarberHome(barber)
				return
			}
		}
	}()
}

// sendBarberHome makes the barber go home
func (shop *BarberShop) sendBarberHome(barber string) {
	fmt.Printf("%s is going home\n", barber)
	shop.BarberhairCutCompleted <- true
}

// cutHair makes the barber cut the given clients hair
func (shop *BarberShop) cutHair(barber string, client string) {
	fmt.Printf("%s is cutting client %s's hair\n", barber, client)
	time.Sleep(shop.hairCutDuraion)
	fmt.Printf("%s is finished cutting client %s's hair\n", barber, client)
}
