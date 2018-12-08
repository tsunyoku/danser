package components

import (
	"github.com/wieku/danser/render/batches"
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/render"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/render/font"
	"github.com/wieku/danser/animation"
	"github.com/wieku/danser/rulesets/osu"
	"github.com/wieku/danser/settings"
)

type Overlay interface {
	Update(int64)
	DrawNormal(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	DrawHUD(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	IsBroken(cursor *render.Cursor) bool
}

/*type knockoutPlayer struct {
	fade      *animation.Glider
	slide     *animation.Glider
	height    *animation.Glider
	lastCombo int64
	hasBroken bool

	lastHit  osu.HitResult
	fadeHit  *animation.Glider
	scaleHit *animation.Glider

	deathFade  *animation.Glider
	deathSlide *animation.Glider
	deathX     float64
}*/

type ScoreOverlay struct {
	//controller *dance.ReplayController
	font       *font.Font
	//players    map[string]*knockoutPlayer
	//names      map[*render.Cursor]string
	lastTime   int64
	//deaths     map[int64]int64
	//generator *rand.Rand
	combo int64
	newCombo int64
	newComboScale *animation.Glider
	ruleset *osu.OsuRuleSet
	cursor *render.Cursor
}

func NewScoreOverlay(ruleset *osu.OsuRuleSet, cursor *render.Cursor) *ScoreOverlay {
	overlay := new(ScoreOverlay)
	//overlay.controller = replayController
	overlay.ruleset = ruleset
	overlay.cursor = cursor
	overlay.font = font.GetFont("Roboto Bold")
	overlay.newComboScale = animation.NewGlider(1)
	/*overlay.players = make(map[string]*knockoutPlayer)
	overlay.names = make(map[*render.Cursor]string)
	overlay.generator = rand.New(rand.NewSource(replayController.GetBeatMap().TimeAdded))*/
	//overlay.deaths = make(map[int64]int64)

	/*for i, r := range replayController.GetReplays() {
		overlay.names[replayController.GetCursors()[i]] = r.Name
		overlay.players[r.Name] = &knockoutPlayer{animation.NewGlider(1), animation.NewGlider(0), animation.NewGlider(settings.Graphics.GetHeightF() * 0.9 * 1.04 / (51)), 0, false, osu.HitResults.Hit300, animation.NewGlider(0), animation.NewGlider(0), animation.NewGlider(0), animation.NewGlider(0), 0}
	}*/
	ruleset.SetListener(func(cursor *render.Cursor, time int64, number int64, position bmath.Vector2d, result osu.HitResult, comboResult osu.ComboResult) {
		/*player := overlay.players[overlay.names[cursor]]

		if result == osu.HitResults.Hit100 || result == osu.HitResults.Hit50 || result == osu.HitResults.Miss {
			player.fadeHit.Reset()
			player.fadeHit.AddEventS(float64(time), float64(time+300), 0.5, 1)
			player.fadeHit.AddEventS(float64(time+600), float64(time+900), 1, 0)
			player.scaleHit.AddEventS(float64(time), float64(time+300), 0.5, 1)
			player.lastHit = result
		}*/

		if comboResult == osu.ComboResults.Increase {
			overlay.combo = overlay.newCombo
			overlay.newCombo++
			overlay.newComboScale.AddEventS(float64(time), float64(time+200), 1.4, 1.0)
		} else if comboResult == osu.ComboResults.Reset {
			overlay.newCombo = 0
			overlay.combo = 0
		}
	})
	return overlay
}

func (overlay *ScoreOverlay) Update(time int64) {
	for sTime := overlay.lastTime + 1; sTime <= time; sTime++ {
		overlay.newComboScale.Update(float64(sTime))
	}
	if overlay.combo != overlay.newCombo && overlay.newComboScale.GetValue() < 1.01 {
		overlay.combo = overlay.newCombo
	}
	overlay.lastTime = time
}

func (overlay *ScoreOverlay) DrawNormal(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	//scl := /*settings.Graphics.GetHeightF() * 0.9*(900.0/1080.0)*/ 384.0*(1080.0/900.0*0.9) / (51)
	//batch.SetScale(1, -1)
	//rescale := /*384.0/512.0 * (1080.0/settings.Graphics.GetHeightF())*/ 1.0
	//for i, r := range overlay.controller.GetReplays() {
	//	player := overlay.players[r.Name]
	//	if player.deathFade.GetValue() >= 0.01 {
	//
	//		batch.SetColor(float64(colors[i].X()), float64(colors[i].Y()), float64(colors[i].Z()), alpha*player.deathFade.GetValue())
	//		width := overlay.font.GetWidth(scl*rescale, r.Name)
	//		overlay.font.Draw(batch, player.deathX-width/2, player.deathSlide.GetValue(), scl*rescale, r.Name)
	//
	//		batch.SetColor(1, 1, 1, alpha*player.deathFade.GetValue())
	//		batch.SetSubScale(scl/2*rescale, scl/2*rescale)
	//		batch.SetTranslation(bmath.NewVec2d(player.deathX+width/2+scl*0.5*rescale, player.deathSlide.GetValue()-scl*0.5*rescale))
	//		batch.DrawUnit(*render.Hit0)
	//	}
	//
	//}
	//batch.SetScale(1, 1)
}

func (overlay *ScoreOverlay) DrawHUD(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	//controller := overlay.controller
	//replays := controller.GetReplays()

	batch.SetColor(1, 1, 1, 0.5)
	overlay.font.Draw(batch, 10, 10, 48*overlay.newComboScale.GetValue(), fmt.Sprintf("%d", overlay.newCombo))
	batch.SetColor(1, 1, 1, 1)
	overlay.font.Draw(batch, 10, 10, 48, fmt.Sprintf("%d", overlay.combo))

	acc, _, score, _ := overlay.ruleset.GetResults(overlay.cursor)

	accText := fmt.Sprintf("%0.2f%%", acc)

	scoreText := fmt.Sprintf("%08d", score)


	overlay.font.Draw(batch, settings.Graphics.GetWidthF()-overlay.font.GetWidth(64, scoreText), settings.Graphics.GetHeightF()-64, 64, scoreText)
	overlay.font.DrawCentered(batch, settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()-32, 32, accText)

	//scl := settings.Graphics.GetHeightF() * 0.9 / 51
	//margin := scl*0.02

	//highestCombo := int64(0)
	//cumulativeHeight := 0.0
	//for _, r := range replays {
	//	cumulativeHeight += overlay.players[r.Name].height.GetValue()
	//	if r.Combo > highestCombo {
	//		highestCombo = r.Combo
	//	}
	//}
	//
	//rowPosY := settings.Graphics.GetHeightF() - (settings.Graphics.GetHeightF()-cumulativeHeight)/2
	//
	//cL := strconv.FormatInt(highestCombo, 10)
	//
	//for i, r := range replays {
	//	player := overlay.players[r.Name]
	//	batch.SetColor(float64(colors[i].X()), float64(colors[i].Y()), float64(colors[i].Z()), alpha*player.fade.GetValue())
	//
	//	rowBaseY := rowPosY - player.height.GetValue()/2 /*+margin*10*/
	//
	//	for j := 0; j < 4; j++ {
	//		if controller.GetClick(i, j) {
	//			batch.SetSubScale(scl*0.9/2, scl*0.9/2)
	//			batch.SetTranslation(bmath.NewVec2d((float64(j)+0.5)*scl, /*rowPosY*/ rowBaseY))
	//			batch.DrawUnit(render.Pixel.GetRegion())
	//		}
	//	}
	//
	//	batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())
	//
	//	accuracy := fmt.Sprintf("%6.2f%% %"+strconv.Itoa(len(cL))+"dx", r.Accuracy, r.Combo)
	//	accuracy1 := "100.00% " + cL + "x "
	//	nWidth := overlay.font.GetWidthMonospaced(scl, accuracy1)
	//
	//	overlay.font.DrawMonospaced(batch, 3*scl, rowBaseY-scl*0.8/2, scl, accuracy)
	//
	//	batch.SetSubScale(scl*0.9/2, -scl*0.9/2)
	//	batch.SetTranslation(bmath.NewVec2d(3*scl+nWidth, rowBaseY))
	//	batch.DrawUnit(*render.GradeTexture[int64(r.Grade)])
	//
	//	batch.SetColor(float64(colors[i].X()), float64(colors[i].Y()), float64(colors[i].Z()), alpha*player.fade.GetValue())
	//	overlay.font.Draw(batch, 4*scl+nWidth, rowBaseY-scl*0.8/2, scl, r.Name)
	//	width := overlay.font.GetWidth(scl, r.Name)
	//
	//	if r.Mods != "" {
	//		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())
	//		overlay.font.Draw(batch, 4*scl+width+nWidth, rowBaseY-scl*0.8/2, scl*0.8, "+"+r.Mods)
	//		width += overlay.font.GetWidth(scl*0.8, "+"+r.Mods)
	//	}
	//
	//	batch.SetColor(1, 1, 1, alpha*player.fade.GetValue()*player.fadeHit.GetValue())
	//	batch.SetSubScale(scl*0.9/2*player.scaleHit.GetValue(), -scl*0.9/2*player.scaleHit.GetValue())
	//	batch.SetTranslation(bmath.NewVec2d(4*scl+width+nWidth+scl*0.5, rowBaseY))
	//
	//	switch player.lastHit {
	//	case osu.HitResults.Hit100:
	//		batch.DrawUnit(*render.Hit100)
	//	case osu.HitResults.Hit50:
	//		batch.DrawUnit(*render.Hit50)
	//	case osu.HitResults.Miss:
	//		batch.DrawUnit(*render.Hit0)
	//	}
	//
	//	rowPosY -= player.height.GetValue()
	//}
}

func (overlay *ScoreOverlay) IsBroken(cursor *render.Cursor) bool {
	return false
}
