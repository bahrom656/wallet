package wallet

import (
	"fmt"
	"github.com/bahrom656/wallet/pkg/types"
	"github.com/google/uuid"
	"reflect"
	"testing"
)

type testService struct {
	*Service
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone:   "992000000001",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

var s Service

func TestFindAccountByID(t *testing.T) {
	s.accounts = []*types.Account{
		{ID: 1, Phone: "+992000000000"},
		{ID: 2, Phone: "+992000000001"},
		{ID: 3, Phone: "+992000000002"},
		{ID: 4, Phone: "+992000000003"},
		{ID: 5, Phone: "+992000000004"},
	}
	acc, err := s.FindAccountByID(4)
	if acc == nil {
		t.Error(err)
	}
}

func TestService_FindPaymentByID_success(t *testing.T) {
	//создаем Сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем найти платёж
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	//сравниваем платежи
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned, error = %v", err)
		return
	}

}
func TestService_FindPaymentByID_fail(t *testing.T) {
	//создаем Сервис
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	// пробуем найти несуществующий платёж
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentByID(): must return error, returned nil")
		return
	}
	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}

}

func TestService_Reject_success(t *testing.T) {
	//создаем Сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	// пробуем отменит платёж
	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}

	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't find payment by id error = %v", savedPayment)
		return
	}

	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't changed payment = %v", err)
		return
	}
	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id error = %v", err)
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't changed, account = %v", savedAccount)
		return
	}
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	//регистрируем там ползователя
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("can't register account, error = %v", err)
	}

	// попалнем его счёт
	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can't deposit account, error = %v", err)
	}

	return account, nil
}
func (s *Service) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	//регистрируем там ползователя
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}

	// попалнем его счёт
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposit account, error = %v", err)
	}

	// выполняем платёжи
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}
	return account, payments, err

}

func TestService_Repeat(t *testing.T) {
	//создаем Сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем найти платёж
	payment := payments[0]
	_, err = s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}

	_, err = s.Repeat(payment.ID)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_FavoritePaymet_success(t *testing.T) {
	//создаем Сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем найти платёж
	payment := payments[0]
	_, err = s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	//добавим в избранный
	_, err = s.FavoritePayment(payment.ID, "Beeline")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_FavoritePaymet_fail(t *testing.T) {
	//создаем Сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем найти платёж
	payment := payments[0]
	_, err = s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	//добавим в избранный
	_, err = s.FavoritePayment(uuid.New().String(), "Beeline")
	if err == nil {
		t.Errorf("error")
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	//создаем Сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем найти платёж
	payment := payments[0]
	_, err = s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	//добавим в избранный
	favpay, err := s.FavoritePayment(payment.ID, "Beeline")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.PayFromFavorite(favpay.ID)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_PayFromFavorite_fail(t *testing.T) {
	//создаем Сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//пробуем найти платёж
	payment := payments[0]
	_, err = s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	//добавим в избранный
	_, err = s.FavoritePayment(payment.ID, "Beeline")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.PayFromFavorite(uuid.New().String())
	if err == nil {
		t.Errorf("error")
		return
	}
}

func TestService_SumPayments(t *testing.T) {
	svc := &Service{}

	for i := 0; i < 103; i++ {
		svc.payments = append(svc.payments, &types.Payment{Amount: 1})
	}

	sum := svc.SumPayments(10)
	if sum != 103 {
		t.Error("incorrect")
	}
}

func Benchmark_SumPayments(b *testing.B) {
	svc := &Service{}

	for i := 0; i < 103; i++ {
		svc.payments = append(svc.payments, &types.Payment{Amount: 1})
	}

	result := 103

	for i := 0; i < b.N; i++ {
		sum := svc.SumPayments(result)
		if result != int(sum) {
			b.Fatalf("invalid result, got %v, want %v", sum, result)
		}
	}
}

func TestService_ExportToFile(t *testing.T) {
	s.accounts = []*types.Account{
		{ID: 1, Phone: "+992000000000"},
		{ID: 2, Phone: "+992000000001"},
		{ID: 3, Phone: "+992000000002"},
		{ID: 4, Phone: "+992000000003"},
		{ID: 5, Phone: "+992000000004"},
	}
	err := s.ExportToFile("export.txt")
	if err != nil {
		t.Error(err)
	}
}
func TestService_ImportFromFile(t *testing.T) {
	err := s.ImportFromFile("export.txt")
	if err != nil {
		t.Error(err)
	}
}

func TestService_Export(t *testing.T) {
	acc, err := s.RegisterAccount("+992981898998")
	if err != nil {
		t.Error(err)
	}
	_, err = s.RegisterAccount("+992981898991")
	if err != nil {
		t.Error(err)
	}
	_, err = s.RegisterAccount("+992981898992")
	if err != nil {
		t.Error(err)
	}
	err = s.Deposit(acc.ID, 5000000)
	if err != nil {
		t.Error(err)
	}
	pay, err := s.Pay(acc.ID, 4050, "auto")
	if err != nil {
		t.Error(err)
	}
	_, err = s.FavoritePayment(pay.ID, "apple")
	if err != nil {
		t.Error(err)
	}
	err = s.Export("data")
	if err != nil {
		t.Error(err)
	}
}
func TestService_Import(t *testing.T) {
	err := s.Import("data")
	if err != nil {
		t.Error(err)
	}
}
func TestService_SumPaymentsWithProgress(t *testing.T) {
	for i := 0; i < 10000; i++ {
		payment := &types.Payment{
			ID:     uuid.New().String(),
			Amount: types.Money(103),
		}
		s.payments = append(s.payments, payment)
	}

	result := s.SumPaymentsWithProgress()
	want := Progress{Result: 1030000, Part: 10000}

	for pro := range result {
		if pro.Result != 1030000 {
			t.Errorf("invalid test: got %v, want %v", pro.Result, want.Result)
		}
	}
}

// func Benchmark_SumPaymentsWithProgress(b *testing.B) {
// 	for i := 0; i < 100000; i++ {
// 		payment := &types.Payment{
// 			ID:     uuid.New().String(),
// 			Amount: types.Money(103),
// 		}
// 		s.payments = append(s.payments, payment)
// 	}
// 	wants := s.SumPaymentsWithProgress()
	

// 	b.ResetTimer()
// 		result := s.SumPaymentsWithProgress()
// 		b.StopTimer()
// 		res := <- result
// 		want := <- wants
// 		want.Result = 10300000
// 		want.Part = 100000
// 			if res.Result != 10300000 && res.Result != 20600000 {
// 				b.Fatalf("invalid result: got %v, want %v", res, want)
// 			}
// 		b.StartTimer()
// }
