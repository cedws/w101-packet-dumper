package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/cedws/w101-client-go/proto"
	"github.com/cedws/w101-proto-go/pkg/aisclient"
	"github.com/cedws/w101-proto-go/pkg/cantrips"
	"github.com/cedws/w101-proto-go/pkg/doodledoug"
	"github.com/cedws/w101-proto-go/pkg/extendedbase"
	"github.com/cedws/w101-proto-go/pkg/game"
	"github.com/cedws/w101-proto-go/pkg/login"
	"github.com/cedws/w101-proto-go/pkg/mg1"
	"github.com/cedws/w101-proto-go/pkg/mg2"
	"github.com/cedws/w101-proto-go/pkg/mg3"
	"github.com/cedws/w101-proto-go/pkg/mg4"
	"github.com/cedws/w101-proto-go/pkg/mg5"
	"github.com/cedws/w101-proto-go/pkg/mg6"
	"github.com/cedws/w101-proto-go/pkg/patch"
	"github.com/cedws/w101-proto-go/pkg/pet"
	"github.com/cedws/w101-proto-go/pkg/quest"
	"github.com/cedws/w101-proto-go/pkg/script"
	"github.com/cedws/w101-proto-go/pkg/skullriders"
	"github.com/cedws/w101-proto-go/pkg/soblocks"
	"github.com/cedws/w101-proto-go/pkg/system"
	"github.com/cedws/w101-proto-go/pkg/testmanager"
	"github.com/cedws/w101-proto-go/pkg/wizard"
	"github.com/cedws/w101-proto-go/pkg/wizardhousing"
)

func main() {
	var (
		inFilename  = flag.String("in", "", "path to the file to read")
		outFilename = flag.String("out", "", "path to the file to write")
	)
	flag.Parse()

	if inFilename == nil || *inFilename == "" {
		flag.Usage()
		os.Exit(1)
	}

	out := os.Stdout

	if outFilename != nil && *outFilename != "" {
		var err error

		out, err = os.Create(*outFilename)
		if err != nil {
			log.Fatalf("error creating output file: %v", err)
		}

		defer out.Close()
	}

	in, err := os.Open(*inFilename)
	if err != nil {
		log.Fatalf("error opening input file: %v", err)
	}
	defer in.Close()

	if err := decodeTo(out, in); err != nil {
		log.Fatalf("error during decoding: %v", err)
	}
}

func decodeTo(w io.Writer, r io.Reader) error {
	frameReader := proto.FrameReader{Reader: r}

	router := proto.NewMessageRouter()
	registerAll(router)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	var messageCount int

	proto.RegisterMiddleware(router, func(message any) {
		type m struct {
			Name    string `json:"name"`
			Message any    `json:"message"`
		}
		enc.Encode(m{
			Name:    reflect.TypeOf(message).Name(),
			Message: message,
		})
		messageCount++
	})

	frameCh := make(chan *proto.Frame, 4)

	go func() {
		for {
			frame, err := frameReader.Read()
			if err != nil {
				break
			}

			if frame.Control {
				continue
			}

			frameCh <- frame
		}

		close(frameCh)
	}()

	start := time.Now()

	for frame := range frameCh {
		var dmlMessage proto.DMLMessage

		if err := dmlMessage.Unmarshal(frame.MessageData); err != nil {
			return fmt.Errorf("error during unmarshal of dml message: %w", err)
		}

		if err := router.Handle(dmlMessage.ServiceID, dmlMessage.OrderNumber, dmlMessage); err != nil {
			return fmt.Errorf("error during handling of dml message (service %d, order %d): %w", dmlMessage.ServiceID, dmlMessage.OrderNumber, err)
		}
	}

	elapsed := time.Since(start)
	messagesPerSec := int(float64(messageCount) / elapsed.Seconds())

	log.Printf("Decoded %d messages in %s (%d messages/sec)\n", messageCount, elapsed, messagesPerSec)

	return nil
}

func registerAll(router *proto.MessageRouter) {
	aisclient.RegisterAisclientService(router, &aisclient.AisclientService{})
	cantrips.RegisterCantripsService(router, &cantrips.CantripsService{})
	doodledoug.RegisterDoodledougService(router, &doodledoug.DoodledougService{})
	extendedbase.RegisterExtendedbaseService(router, &extendedbase.ExtendedbaseService{})
	game.RegisterGameService(router, &game.GameService{})
	login.RegisterLoginService(router, &login.LoginService{})
	mg1.RegisterMg1Service(router, &mg1.Mg1Service{})
	mg2.RegisterMg2Service(router, &mg2.Mg2Service{})
	mg3.RegisterMg3Service(router, &mg3.Mg3Service{})
	mg4.RegisterMg4Service(router, &mg4.Mg4Service{})
	mg5.RegisterMg5Service(router, &mg5.Mg5Service{})
	mg6.RegisterMg6Service(router, &mg6.Mg6Service{})
	patch.RegisterPatchService(router, &patch.PatchService{})
	pet.RegisterPetService(router, &pet.PetService{})
	quest.RegisterQuestService(router, &quest.QuestService{})
	script.RegisterScriptService(router, &script.ScriptService{})
	skullriders.RegisterSkullridersService(router, &skullriders.SkullridersService{})
	soblocks.RegisterSoblocksService(router, &soblocks.SoblocksService{})
	system.RegisterSystemService(router, &system.SystemService{})
	testmanager.RegisterTestmanagerService(router, &testmanager.TestmanagerService{})
	wizard.RegisterWizardService(router, &wizard.WizardService{})
	wizardhousing.RegisterWizardhousingService(router, &wizardhousing.WizardhousingService{})
}
