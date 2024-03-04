// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-02-28 16:16:02

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"encoding/json"
	"fmt"

	"github.com/yuansl/playground/util"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
)

func main() {
	raw := []byte(`[{"t": 1708959900, "v": 23394607}, {"t": 1708968900, "v": 23548608}, {"t": 1708962000, "v": 22291302}, {"t": 1708988700, "v": 27561726}, {"t": 1708982100, "v": 24446864}, {"t": 1708941000, "v": 23993267}, {"t": 1708936800, "v": 23882780}, {"t": 1708913700, "v": 24436950}, {"t": 1708977600, "v": 24575880}, {"t": 1708928400, "v": 24193486}, {"t": 1708975500, "v": 24178539}, {"t": 1708906500, "v": 24191941}, {"t": 1708962300, "v": 22810842}, {"t": 1708986900, "v": 27563912}, {"t": 1708988400, "v": 27487372}, {"t": 1708915500, "v": 24455275}, {"t": 1708912500, "v": 23856861}, {"t": 1708985400, "v": 27512700}, {"t": 1708968000, "v": 23571082}, {"t": 1708932000, "v": 24033176}, {"t": 1708912800, "v": 24252594}, {"t": 1708962900, "v": 22563616}, {"t": 1708935000, "v": 23719115}, {"t": 1708955400, "v": 23734636}, {"t": 1708967400, "v": 23925220}, {"t": 1708947600, "v": 23555014}, {"t": 1708976400, "v": 24176831}, {"t": 1708958400, "v": 22995796}, {"t": 1708964400, "v": 22999286}, {"t": 1708945500, "v": 24028822}, {"t": 1708921500, "v": 24065965}, {"t": 1708911600, "v": 24145263}, {"t": 1708940700, "v": 24056176}, {"t": 1708932300, "v": 23745583}, {"t": 1708917300, "v": 24052927}, {"t": 1708976100, "v": 24155702}, {"t": 1708905900, "v": 24426248}, {"t": 1708970700, "v": 23774362}, {"t": 1708966800, "v": 23717919}, {"t": 1708981200, "v": 24299680}, {"t": 1708920300, "v": 24281626}, {"t": 1708973700, "v": 24214396}, {"t": 1708989000, "v": 27516268}, {"t": 1708921200, "v": 24062468}, {"t": 1708952700, "v": 24196879}, {"t": 1708931700, "v": 23808224}, {"t": 1708991400, "v": 27527178}, {"t": 1708959600, "v": 23233410}, {"t": 1708954500, "v": 23715773}, {"t": 1708955100, "v": 23800615}, {"t": 1708983000, "v": 24016735}, {"t": 1708926300, "v": 24316288}, {"t": 1708948500, "v": 24056085}, {"t": 1708936500, "v": 24189189}, {"t": 1708907400, "v": 24702222}, {"t": 1708943400, "v": 24288908}, {"t": 1708924500, "v": 23912882}, {"t": 1708907700, "v": 24465677}, {"t": 1708943100, "v": 24350238}, {"t": 1708957500, "v": 23116020}, {"t": 1708987200, "v": 27422372}, {"t": 1708953000, "v": 24262887}, {"t": 1708944900, "v": 24091995}, {"t": 1708942200, "v": 24289194}, {"t": 1708935900, "v": 23487365}, {"t": 1708975200, "v": 24069616}, {"t": 1708931400, "v": 24091244}, {"t": 1708966200, "v": 23665465}, {"t": 1708984800, "v": 27503564}, {"t": 1708987500, "v": 27427838}, {"t": 1708923300, "v": 24231845}, {"t": 1708914000, "v": 24586988}, {"t": 1708918500, "v": 24066703}, {"t": 1708935600, "v": 23889079}, {"t": 1708914600, "v": 24159661}, {"t": 1708932600, "v": 23729575}, {"t": 1708934100, "v": 23627220}, {"t": 1708964100, "v": 22969445}, {"t": 1708963800, "v": 23061599}, {"t": 1708919100, "v": 24280485}, {"t": 1708956600, "v": 23579182}, {"t": 1708961400, "v": 23202125}, {"t": 1708922400, "v": 24128557}, {"t": 1708939800, "v": 24126676}, {"t": 1708974600, "v": 24068624}, {"t": 1708977300, "v": 24044180}, {"t": 1708909200, "v": 24463242}, {"t": 1708938900, "v": 23844635}, {"t": 1708972800, "v": 24125182}, {"t": 1708925100, "v": 23845136}, {"t": 1708972500, "v": 24004087}, {"t": 1708972200, "v": 23952976}, {"t": 1708940400, "v": 24231975}, {"t": 1708954200, "v": 24284440}, {"t": 1708941600, "v": 24000425}, {"t": 1708909500, "v": 24218084}, {"t": 1708976700, "v": 24570624}, {"t": 1708950600, "v": 23900162}, {"t": 1708957800, "v": 23152614}, {"t": 1708938600, "v": 24390029}, {"t": 1708933500, "v": 24054545}, {"t": 1708938000, "v": 23851797}, {"t": 1708953600, "v": 23939454}, {"t": 1708967700, "v": 23210795}, {"t": 1708930800, "v": 20682077}, {"t": 1708959300, "v": 22823981}, {"t": 1708973400, "v": 24053306}, {"t": 1708911300, "v": 24290486}, {"t": 1708913100, "v": 24272394}, {"t": 1708989600, "v": 27560417}, {"t": 1708913400, "v": 24257056}, {"t": 1708958100, "v": 23009760}, {"t": 1708929300, "v": 23981855}, {"t": 1708925700, "v": 24340588}, {"t": 1708923000, "v": 24193415}, {"t": 1708982700, "v": 24112875}, {"t": 1708917900, "v": 24180851}, {"t": 1708956300, "v": 24029848}, {"t": 1708965000, "v": 22613130}, {"t": 1708990500, "v": 27522783}, {"t": 1708978800, "v": 24419819}, {"t": 1708969500, "v": 23716429}, {"t": 1708955700, "v": 23401762}, {"t": 1708990800, "v": 27553413}, {"t": 1708969200, "v": 23362012}, {"t": 1708935300, "v": 24024851}, {"t": 1708949100, "v": 24131427}, {"t": 1708984200, "v": 27589616}, {"t": 1708979400, "v": 24340276}, {"t": 1708965600, "v": 22555700}, {"t": 1708980300, "v": 24501613}, {"t": 1708954800, "v": 24371968}, {"t": 1708933200, "v": 24096818}, {"t": 1708930500, "v": 23632822}, {"t": 1708936200, "v": 24304175}, {"t": 1708986600, "v": 27578330}, {"t": 1708910400, "v": 24291500}, {"t": 1708929900, "v": 23965654}, {"t": 1708926000, "v": 23594199}, {"t": 1708983300, "v": 24264396}, {"t": 1708910700, "v": 24323220}, {"t": 1708922700, "v": 24078040}, {"t": 1708946400, "v": 24017312}, {"t": 1708962600, "v": 23049908}, {"t": 1708977900, "v": 23949092}, {"t": 1708944000, "v": 22326466}, {"t": 1708950900, "v": 23824404}, {"t": 1708915800, "v": 24287604}, {"t": 1708917600, "v": 24433824}, {"t": 1708922100, "v": 24113654}, {"t": 1708974000, "v": 23800267}, {"t": 1708964700, "v": 22493260}, {"t": 1708963500, "v": 23056267}, {"t": 1708944600, "v": 24607517}, {"t": 1708937700, "v": 23573372}, {"t": 1708985700, "v": 27641724}, {"t": 1708991700, "v": 27479136}, {"t": 1708914300, "v": 23858779}, {"t": 1708948200, "v": 24313761}, {"t": 1708925400, "v": 23916751}, {"t": 1708940100, "v": 24546066}, {"t": 1708981500, "v": 24542404}, {"t": 1708974300, "v": 23991641}, {"t": 1708939500, "v": 23817363}, {"t": 1708945200, "v": 24144958}, {"t": 1708908000, "v": 24248458}, {"t": 1708985100, "v": 27579217}, {"t": 1708977000, "v": 24236809}, {"t": 1708924200, "v": 24101539}, {"t": 1708989900, "v": 27553439}, {"t": 1708949400, "v": 24182253}, {"t": 1708980600, "v": 24089579}, {"t": 1708950300, "v": 24153734}, {"t": 1708956900, "v": 22856261}, {"t": 1708927800, "v": 23596921}, {"t": 1708981800, "v": 24101904}, {"t": 1708978500, "v": 24114658}, {"t": 1708957200, "v": 22692870}, {"t": 1708920900, "v": 24386262}, {"t": 1708929000, "v": 23738945}, {"t": 1708937400, "v": 23807080}, {"t": 1708916100, "v": 24351458}, {"t": 1708967100, "v": 23478160}, {"t": 1708906800, "v": 24396023}, {"t": 1708941300, "v": 24055852}, {"t": 1708942500, "v": 24320025}, {"t": 1708959000, "v": 22649977}, {"t": 1708971600, "v": 24083462}, {"t": 1708983900, "v": 27564799}, {"t": 1708905600, "v": 24488875}, {"t": 1708949700, "v": 23869689}, {"t": 1708991100, "v": 27590623}, {"t": 1708966500, "v": 23229545}, {"t": 1708907100, "v": 24669039}, {"t": 1708947900, "v": 23968848}, {"t": 1708945800, "v": 24145690}, {"t": 1708917000, "v": 24268850}, {"t": 1708918200, "v": 23997510}, {"t": 1708927500, "v": 24340735}, {"t": 1708909800, "v": 24295659}, {"t": 1708914900, "v": 24277648}, {"t": 1708948800, "v": 24139468}, {"t": 1708986300, "v": 27596230}, {"t": 1708947000, "v": 24057662}, {"t": 1708929600, "v": 24108394}, {"t": 1708951800, "v": 24296422}, {"t": 1708969800, "v": 23809333}, {"t": 1708937100, "v": 23914110}, {"t": 1708911000, "v": 24167504}, {"t": 1708970100, "v": 24157191}, {"t": 1708980000, "v": 23972504}, {"t": 1708912200, "v": 24367235}, {"t": 1708942800, "v": 23588519}, {"t": 1708968600, "v": 23331904}, {"t": 1708968300, "v": 23517703}, {"t": 1708978200, "v": 24149237}, {"t": 1708961100, "v": 22921106}, {"t": 1708950000, "v": 23865248}, {"t": 1708916400, "v": 24287534}, {"t": 1708990200, "v": 27590611}, {"t": 1708980900, "v": 24611304}, {"t": 1708920000, "v": 24073591}, {"t": 1708910100, "v": 24307033}, {"t": 1708934400, "v": 24269551}, {"t": 1708918800, "v": 24233048}, {"t": 1708915200, "v": 24174828}, {"t": 1708926900, "v": 23658822}, {"t": 1708960800, "v": 22591307}, {"t": 1708987800, "v": 27491292}, {"t": 1708979100, "v": 24203695}, {"t": 1708921800, "v": 24086916}, {"t": 1708953300, "v": 24035729}, {"t": 1708943700, "v": 23766728}, {"t": 1708952400, "v": 24086306}, {"t": 1708919400, "v": 24274766}, {"t": 1708971300, "v": 24011483}, {"t": 1708908900, "v": 24138422}, {"t": 1708947300, "v": 24307617}, {"t": 1708946100, "v": 24284914}, {"t": 1708933800, "v": 23626858}, {"t": 1708924800, "v": 23958635}, {"t": 1708984500, "v": 27605915}, {"t": 1708960200, "v": 22936226}, {"t": 1708963200, "v": 22842795}, {"t": 1708930200, "v": 23954491}, {"t": 1708932900, "v": 24006457}, {"t": 1708939200, "v": 24287105}, {"t": 1708965300, "v": 22797847}, {"t": 1708971900, "v": 23536523}, {"t": 1708906200, "v": 24538846}, {"t": 1708916700, "v": 24228249}, {"t": 1708928100, "v": 24463290}, {"t": 1708938300, "v": 24120781}, {"t": 1708956000, "v": 23400736}, {"t": 1708944300, "v": 14647459}, {"t": 1708908300, "v": 24519747}, {"t": 1708953900, "v": 23762679}, {"t": 1708923600, "v": 23836216}, {"t": 1708961700, "v": 23217828}, {"t": 1708983600, "v": 24871237}, {"t": 1708971000, "v": 23980271}, {"t": 1708958700, "v": 23146920}, {"t": 1708975800, "v": 24371049}, {"t": 1708951200, "v": 24238094}, {"t": 1708934700, "v": 23674320}, {"t": 1708923900, "v": 24157394}, {"t": 1708979700, "v": 24995572}, {"t": 1708951500, "v": 23581887}, {"t": 1708946700, "v": 24179726}, {"t": 1708920600, "v": 23863392}, {"t": 1708960500, "v": 22925713}, {"t": 1708988100, "v": 27425688}, {"t": 1708911900, "v": 24364783}, {"t": 1708928700, "v": 23814057}, {"t": 1708986000, "v": 27609847}, {"t": 1708941900, "v": 24475672}, {"t": 1708965900, "v": 23052204}, {"t": 1708973100, "v": 24565233}, {"t": 1708919700, "v": 24650104}, {"t": 1708908600, "v": 24551751}, {"t": 1708926600, "v": 24351130}, {"t": 1708927200, "v": 23613654}, {"t": 1708931100, "v": 23539852}, {"t": 1708952100, "v": 24751469}, {"t": 1708989300, "v": 27551295}, {"t": 1708982400, "v": 24471493}, {"t": 1708970400, "v": 23650231}, {"t": 1708974900, "v": 24381561}]`)
	var timeseries []struct {
		Time  int64 `json:"t"`
		Value int64 `json:"v"`
	}
	err := json.Unmarshal(raw, &timeseries)
	if err != nil {
		util.Fatal("json.Unmarshal:", err)
	}
	var traffic int64
	for _, ts := range timeseries {
		traffic += ts.Value
	}
	fmt.Printf("total traffic: %d GiB\n", traffic/GiB)
}