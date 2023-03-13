package observer

// EventRunner interfejs očekuje od objekta da definise ponašanje nakon
// primanja obaveštenja od Observer objekta
type EventRunner interface{
	Trigger(Message) error
}

// Message interfejs treba da obezbedi EventRunner objektima informacije od objekta 
// koji se nadgleda. Pozeljno je da objekat poruke bude podudaran sa XML ili JSON marshaler interfejsima
// od metode Format se ocekuje da vrati byte slajs koji sadrži oblik poruke u tekstualnom formatu.
// formatType parametar definiše koja Marshal funkcija treba da se pozove
type Message interface{
	Format(formatType string) []byte
}
// Interfejs SelfWatcher zahteva da objekat implementira logiku kojom će se 
// izvršavati nadgledanje metoda WatchYourself bi na odvojenoj gorutini trebala da
// šalje signale niz kanal kada se ispune zahtevi za slanje
// očekuje se da metoda WatchYourself zatvori kanal, logiku zatvaranja kanala 
// treba da implementira korisnik
type SelfWatcher interface{
	WatchYourself(communicate chan Message, err chan error)
}

// Observer Objekat nadgleda Target objekat, kada Target prijavi promenu, observer šalje svim EventRunner objektima
// signal za izvršavanje potrebnih događaja
type Observer struct {
	Observers []EventRunner
	Target    SelfWatcher
}


// Notify operacija šalje signal svim EventRunnerima da pokrenu potrebne događaje
func (o *Observer) Notify(m Message){
	for _, observer := range o.Observers{
		observer.Trigger(m)
	}
}

// Watch funkcija pokreće mehanizam kojim se ciljani objekat nadgleda, i definiše kanal preko koga se primaju signali od objekta
// po primanju signala objekat poziva funkciju notify, u slučaju da dođe do grške prilikom nadgledanja
// proverava se da li su kanali lepo zatvoreni, greška bi trebala da se pošalje nazad po errs 
func (o *Observer) Watch() error{
	messages := make(chan Message, 1)
	errs := make(chan error, 1)
	go o.Target.WatchYourself(messages , errs)

	for {
		select{
		case message, ok := <-messages:
			if ok{
				o.Notify(message)
			} 
		case err := <-errs:
			// Proveri da li su kanali signal i errs lepo zatvoreni po izlazu ako, ako ne zatvori ih
			_, ok := <-messages
			if !ok {
				close(messages)
			}
			if _, ok := <-errs; !ok {
				close(errs)
			}

			return err
		}
		if _, ok := <-messages; !ok {
			break;
		}
	}
	return nil
}

func (o *Observer) AddObserver(e EventRunner) {
	o.Observers = append(o.Observers, e)
}

func New(target SelfWatcher) *Observer{
	return &Observer{
		Observers:  make([]EventRunner, 0),
		Target: target,
	}
}