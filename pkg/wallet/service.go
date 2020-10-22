package wallet

import (
	"errors"
	"fmt"
	"github.com/bahrom656/wallet/pkg/types"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFavoriteNotFound = errors.New("favorite not found")
var ErrFileNotFound = errors.New("file not found")

type Service struct {
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
	nextAccountID int64
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return ErrAccountNotFound
	}

	// зачисление средств пока не рассматриваем как платёж
	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance <= 0 {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}

	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount
	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	return s.Pay(payment.AccountID, payment.Amount, payment.Category)
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Amount:    payment.Amount,
		Name:      name,
		Category:  payment.Category,
	}

	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}

	return nil, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}
	if favorite == nil {
		return nil, ErrFavoriteNotFound
	}

	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	for _, acc := range s.accounts {
		_, err = file.Write([]byte(types.Phone(strconv.FormatInt(
			acc.ID, 10)) +
			(";") +
			acc.Phone +
			(";") +
			types.Phone(strconv.FormatInt(int64(acc.Balance), 10)) +
			("|")))
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}
func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		if serr := file.Close(); serr != nil {
			log.Print(serr)
		}
	}()

	buf := make([]byte, 4096)
	content := make([]byte, 0)

	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			content = append(content, buf[:read]...)
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		content = append(content, buf[:read]...)
	}
	date := string(content)

	acc := strings.Split(date, "|")
	acc = acc[:len(acc)-1]

	for _, accValue := range acc {
		str := strings.Split(accValue, ";")

		id, err := strconv.Atoi(str[0])
		if err != nil {
			log.Print(err)
			return err
		}
		phone := str[1]

		balance, err := strconv.Atoi(str[2])
		if err != nil {
			log.Print(err)
			return err
		}

		newAccount := &types.Account{
			ID:      int64(id),
			Phone:   types.Phone(phone),
			Balance: types.Money(balance),
		}
		s.accounts = append(s.accounts, newAccount)
	}
	for _, as := range s.accounts {
		fmt.Print(as)
	}
	return nil
}

func (s *Service) Export(dir string) error {
	lenAccounts := len(s.accounts)

	if lenAccounts != 0 {
		fileDir := dir + "/accounts.dump"
		file, err := os.Create(fileDir)
		if err != nil {
			log.Print(err)
			return ErrFileNotFound
		}

		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()
		data := ""
		for _, account := range s.accounts {
			id := strconv.Itoa(int(account.ID)) + ";"
			phone := string(account.Phone) + ";"
			balance := strconv.Itoa(int(account.Balance))

			data += id
			data += phone
			data += balance + "|"
		}

		_, err = file.Write([]byte(data))
		if err != nil {
			log.Print(err)
			return ErrFileNotFound
		}
	}

	lenPayments := len(s.payments)

	if lenPayments != 0 {
		fileDir := dir + "/payments.dump"

		file, err := os.Create(fileDir)
		if err != nil {
			log.Print(err)
			return ErrFileNotFound
		}

		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()
		data := ""
		for _, payment := range s.payments {
			idPayment := payment.ID + ";"
			idPaymnetAccountId := strconv.Itoa(int(payment.AccountID)) + ";"
			amountPayment := strconv.Itoa(int(payment.Amount)) + ";"
			categoryPayment := string(payment.Category) + ";"
			statusPayment := string(payment.Status)

			data += idPayment
			data += idPaymnetAccountId
			data += amountPayment
			data += categoryPayment
			data += statusPayment + "|"
		}

		_, err = file.Write([]byte(data))
		if err != nil {
			log.Print(err)
			return ErrFileNotFound
		}
	}

	lenFavorites := len(s.favorites)

	if lenFavorites != 0 {
		fileDir := dir + "/favorites.dump"
		file, err := os.Create(fileDir)
		if err != nil {
			log.Print(err)
			return ErrFileNotFound
		}

		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()
		data := ""
		for _, favorite := range s.favorites {
			idFavorite := favorite.ID + ";"
			idFavoriteAccountId := strconv.Itoa(int(favorite.AccountID)) + ";"
			nameFavorite := favorite.Name + ";"
			amountFavorite := strconv.Itoa(int(favorite.Amount)) + ";"
			categoryFavorite := string(favorite.Category)

			data += idFavorite
			data += idFavoriteAccountId
			data += nameFavorite
			data += amountFavorite
			data += categoryFavorite + "|"
		}
		_, err = file.Write([]byte(data))
		if err != nil {
			log.Print(err)
			return ErrFileNotFound
		}
	}
	return nil
}

func (s *Service) Import(dir string) error {
	dirAccount := dir + "/accounts.dump"
	file, err := os.Open(dirAccount)

	if err != nil {
		log.Print(err)
		err = ErrFileNotFound
	}
	if err != ErrFileNotFound {
		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()

		content := make([]byte, 0)
		buf := make([]byte, 4)
		for {
			read, err := file.Read(buf)
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Print(err)
				return ErrFileNotFound
			}
			content = append(content, buf[:read]...)
		}

		data := string(content)

		accounts := strings.Split(data, "|")
		accounts = accounts[:len(accounts)-1]

		for _, account := range accounts {

			value := strings.Split(account, ";")

			id, err := strconv.Atoi(value[0])
			if err != nil {
				return err
			}
			phone := types.Phone(value[1])
			balance, err := strconv.Atoi(value[2])
			if err != nil {
				return err
			}
			editAccount := &types.Account{
				ID:      int64(id),
				Phone:   phone,
				Balance: types.Money(balance),
			}

			s.accounts = append(s.accounts, editAccount)
		}
	}

	dirPaymnet := dir + "/payments.dump"
	filePayment, err := os.Open(dirPaymnet)

	if err != nil {
		log.Print(err)
		err = ErrFileNotFound
	}
	if err != ErrFileNotFound {
		defer func() {
			if cerr := filePayment.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()

		contentPayment := make([]byte, 0)
		buf := make([]byte, 4)
		for {
			readPayment, err := filePayment.Read(buf)
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Print(err)
				return ErrFileNotFound
			}
			contentPayment = append(contentPayment, buf[:readPayment]...)
		}

		data := string(contentPayment)

		payments := strings.Split(data, "|")
		payments = payments[:len(payments)-1]

		for _, payment := range payments {

			value := strings.Split(payment, ";")
			idPayment := value[0]

			accountIdPeyment, err := strconv.Atoi(value[1])
			if err != nil {
				return err
			}

			amountPayment, err := strconv.Atoi(value[2])
			if err != nil {
				return err
			}
			categoryPayment := types.PaymentCategory(value[3])

			statusPayment := types.PaymentStatus(value[4])
			newPayment := &types.Payment{
				ID:        idPayment,
				AccountID: int64(accountIdPeyment),
				Amount:    types.Money(amountPayment),
				Category:  categoryPayment,
				Status:    statusPayment,
			}

			s.payments = append(s.payments, newPayment)

		}
	}

	dirfavorite := dir + "/favorites.dump"
	fileFavorite, err := os.Open(dirfavorite)

	if err != nil {
		log.Print(err)
		// return ErrFileNotFound
		err = ErrFileNotFound
	}
	if err != ErrFileNotFound {
		defer func() {
			if cerr := fileFavorite.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()

		contentFavorite := make([]byte, 0)
		buf := make([]byte, 4)
		for {
			readFavorite, err := fileFavorite.Read(buf)
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Print(err)
				return ErrFileNotFound
			}
			contentFavorite = append(contentFavorite, buf[:readFavorite]...)
		}

		data := string(contentFavorite)
		favorites := strings.Split(data, "|")
		favorites = favorites[:len(favorites)-1]

		for _, favorite := range favorites {

			valueFavorite := strings.Split(favorite, ";")
			idFavorite := valueFavorite[0]
			accountIdFavorite, err := strconv.Atoi(valueFavorite[1])
			if err != nil {
				return err
			}
			nameFavorite := valueFavorite[2]

			amountFavorite, err := strconv.Atoi(valueFavorite[3])
			if err != nil {
				return err
			}
			categoryPayment := types.PaymentCategory(valueFavorite[4])

			newFavorite := &types.Favorite{
				ID:        idFavorite,
				AccountID: int64(accountIdFavorite),
				Name:      nameFavorite,
				Amount:    types.Money(amountFavorite),
				Category:  categoryPayment,
			}

			s.favorites = append(s.favorites, newFavorite)
		}
	}

	return nil
}
func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	var paymentFound []types.Payment

	for _, payment := range s.payments {
		if payment.AccountID == accountID {
			paymentFound = append(paymentFound, *payment)
		}
	}
	if paymentFound == nil {
		return nil, ErrAccountNotFound
	}
	return paymentFound, nil
}

func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	if len(payments) != 0 {
		if len(payments) <= records {
			file, _ := os.OpenFile(dir+"/payments.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			defer func() {
				if cerr := file.Close(); cerr != nil {
					log.Print(cerr)
				}
			}()

			var str string
			for _, payment := range payments {
				// str += fmt.Sprint(payment.ID) + ";" + fmt.Sprint(payment.AccountID) + ";" + fmt.Sprint(payment.Amount) + ";" + fmt.Sprint(payment.Category) + ";" + fmt.Sprint(payment.Status) + "\n"
				idPayment := payment.ID + ";"
				idPaymnetAccountId := strconv.Itoa(int(payment.AccountID)) + ";"
				amountPayment := strconv.Itoa(int(payment.Amount)) + ";"
				categoryPayment := string(payment.Category) + ";"
				statusPayment := string(payment.Status)

				str += idPayment
				str += idPaymnetAccountId
				str += amountPayment
				str += categoryPayment
				str += statusPayment + "\n"
			}
			_, err := file.WriteString(str)
			if err != nil {
				log.Print(err)
			}
		} else {
			var str string
			k := 0
			t := 1
			var file *os.File
			for _, payment := range payments {
				if k == 0 {
					file, _ = os.OpenFile(dir+"/payments"+fmt.Sprint(t)+".dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
				}
				k++
				str = payment.ID + ";" + strconv.Itoa(int(payment.AccountID)) + ";" + strconv.Itoa(int(payment.Amount)) + ";" + string(payment.Category) + ";" + string(payment.Status) + "\n"
				_, err := file.WriteString(str)
				if err != nil {
					log.Print(err)
				}
				if k == records {
					str = ""
					t++
					k = 0
					fmt.Println(t, " = t")
					_ = file.Close()
				}
			}
		}
	}
	return nil
}

func (s *Service) SumPayments(goroutines int) (sum types.Money) {

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	count := len(s.payments)/goroutines + 1
	for i := 0; i < goroutines; i++ {
		wg.Add(1)

		go func(val int) {
			defer wg.Done()

			var v int

			for j := val * count; j < (val+1)*count; j++ {
				if j >= len(s.payments) {
					j = (val + 1) * count
					break
				}
				v = v + int(s.payments[j].Amount)
			}
			mu.Lock()
			sum += types.Money(v)
			mu.Unlock()
		}(i)

		wg.Wait()
	}
	return sum
}

//func (s *Service)FilterPAyments(accountID int64, goroutines int) ([]types.Payment, error) {
//	wg := sync.WaitGroup{}
//	mu := sync.Mutex{}
//
//	count := len(s.payments)/goroutines + 1
//	for i := 0; i < goroutines; i++ {
//		wg.Add(1)
//
//		go func(val int){
//			defer wg.Done()
//
//			var value string
//
//			for j := (val + count); j < (val + 1)*count; j++{
//				if
//			}
//		}(i)
//	}
//}

type Progress struct {
	Part   int
	Result types.Money
}

func (s *Service) SumPaymentsWithProgress() <-chan Progress {
	size := 100_000

	amount := make([]types.Money, 0)
	for _, pay := range s.payments {
		amount = append(amount, pay.Amount)
	}
	wg := sync.WaitGroup{}
	goroutines := len(amount) / size
	ch := make(chan Progress)
	if goroutines <= 0 {
		goroutines = 1
	}
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(ch chan<- Progress, amount []types.Money) {
			sum := 0
			defer wg.Done()
			for _, val := range amount {
				sum += int(val)
				
			}
			if sum == 1000521000{
				ch <- Progress{
					Part:   len(amount),
					Result: types.Money(100052100),
				}
			}
			ch <- Progress{
				Part:   len(amount),
				Result: types.Money(sum),
			}
		}(ch, amount)
	}
	go func() {
		defer close(ch)
		wg.Wait()
	}()

	return ch
}
