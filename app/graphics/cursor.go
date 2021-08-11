package graphics

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tsunyoku/danser/app/bmath"
	"github.com/tsunyoku/danser/app/bmath/camera"
	"github.com/tsunyoku/danser/app/settings"
	"github.com/tsunyoku/danser/app/skin"
	"github.com/tsunyoku/danser/framework/graphics/batch"
	"github.com/tsunyoku/danser/framework/graphics/blend"
	"github.com/tsunyoku/danser/framework/graphics/buffer"
	"github.com/tsunyoku/danser/framework/graphics/sprite"
	"github.com/tsunyoku/danser/framework/graphics/texture"
	"github.com/tsunyoku/danser/framework/math/animation"
	"github.com/tsunyoku/danser/framework/math/animation/easing"
	color2 "github.com/tsunyoku/danser/framework/math/color"
	"github.com/tsunyoku/danser/framework/math/math32"
	"github.com/tsunyoku/danser/framework/math/vector"
	"math"
	"math/rand"
	"time"
)

type cursorRenderer interface {
	Update(delta float64, position vector.Vector2f)
	UpdateRenderer()
	DrawM(scale, expand float64, batch *batch.QuadBatch, color color2.Color, colorGlow color2.Color)
}

var cursorFbo *buffer.Framebuffer = nil
var cursorFBOSprite *sprite.Sprite

var cursorSpaceFbo *buffer.Framebuffer = nil
var cursorSpaceFBOSprite *sprite.Sprite

var fboBatch *batch.QuadBatch

var Camera *camera.Camera
var osuRect camera.Rectangle

var useAdditive = false

func initCursor() {
	if settings.Cursor.TrailStyle < 1 || settings.Cursor.TrailStyle > 4 {
		panic("Wrong cursor trail type")
	}

	cursorFbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	region := cursorFbo.Texture().GetRegion()
	cursorFBOSprite = sprite.NewSpriteSingle(&region, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), bmath.Origin.Centre)

	cursorSpaceFbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	regionSpace := cursorSpaceFbo.Texture().GetRegion()
	cursorSpaceFBOSprite = sprite.NewSpriteSingle(&regionSpace, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), bmath.Origin.Centre)

	fboBatch = batch.NewQuadBatchSize(1)
	fboBatch.SetCamera(mgl32.Ortho(0, float32(settings.Graphics.GetWidth()), 0, float32(settings.Graphics.GetHeight()), -1, 1))

	osuRect = Camera.GetWorldRect()
}

type Cursor struct {
	scale *animation.Glider

	lastLeftState, lastRightState bool

	LeftButton, RightButton bool
	LeftKey, RightKey       bool
	LeftMouse, RightMouse   bool

	IsReplayFrame bool // TODO: temporary hacky solution for spinners
	IsPlayer      bool
	IsAutoplay    bool

	OldSpinnerScoring bool

	LastFrameTime    int64 //
	CurrentFrameTime int64 //
	RawPosition      vector.Vector2f
	InvertDisplay    bool

	Position vector.Vector2f

	Name      string
	ScoreID   int64
	ScoreTime time.Time

	lastSetting bool

	renderer cursorRenderer

	SmokeKey           bool
	lastSmokeKey       bool
	smokePointCount    int
	firstSmokePosition vector.Vector2f
	lastSmokePosition  vector.Vector2f
	smokeTexture       *texture.TextureRegion
	smokeContainer     *sprite.SpriteManager

	rippleContainer *sprite.SpriteManager
	time            float64
}

func NewCursor() *Cursor {
	if cursorFbo == nil {
		initCursor()
	}

	cursor := &Cursor{Position: vector.NewVec2f(100, 100)}
	cursor.scale = animation.NewGlider(1.0)

	cursor.lastSetting = settings.Skin.Cursor.UseSkinCursor

	if cursor.lastSetting {
		cursor.renderer = newOsuRenderer()
	} else {
		cursor.renderer = newDanserRenderer()
	}

	cursor.smokeTexture = skin.GetTexture("cursor-smoke")
	cursor.smokeContainer = sprite.NewSpriteManager()

	cursor.rippleContainer = sprite.NewSpriteManager()

	return cursor
}

func (cursor *Cursor) SetPos(pt vector.Vector2f) {
	cursor.RawPosition = pt
	tmp := pt

	if cursor.InvertDisplay {
		tmp.Y = 384 - tmp.Y
	}

	if settings.Cursor.BounceOnEdges && settings.DIVIDES <= 2 {
		tmp.X -= osuRect.MinX
		tmp.Y -= osuRect.MinY
		tmp.X = math32.Mod(tmp.X, 2*(osuRect.MaxX-osuRect.MinX))
		tmp.Y = math32.Mod(tmp.Y, 2*(osuRect.MaxY-osuRect.MinY))
		tmp.X += osuRect.MinX
		tmp.Y += osuRect.MinY

		for {
			ok1, ok2 := false, false

			if tmp.X < osuRect.MinX {
				tmp.X = 2*osuRect.MinX - tmp.X
			} else if tmp.X > osuRect.MaxX {
				tmp.X = 2*osuRect.MaxX - tmp.X
			} else {
				ok1 = true
			}

			if tmp.Y < osuRect.MinY {
				tmp.Y = 2*osuRect.MinY - tmp.Y
			} else if tmp.Y > osuRect.MaxY {
				tmp.Y = 2*osuRect.MaxY - tmp.Y
			} else {
				ok2 = true
			}

			if ok1 && ok2 {
				break
			}
		}
	}

	cursor.Position = tmp
}

func (cursor *Cursor) SetScreenPos(pt vector.Vector2f) {
	cursor.SetPos(Camera.Unproject(pt.Copy64()).Copy32())
}

func (cursor *Cursor) Update(delta float64) {
	delta = math.Abs(delta)
	cursor.time += delta

	leftState := cursor.LeftKey || cursor.LeftMouse
	rightState := cursor.RightKey || cursor.RightMouse

	if settings.Cursor.CursorRipples && settings.PLAYERS == 1 && ((!cursor.lastLeftState && leftState) || (!cursor.lastRightState && rightState)) {
		spr := sprite.NewSpriteSingle(skin.GetTextureSource("ripple", skin.LOCAL), cursor.time, cursor.Position.Copy64(), bmath.Origin.Centre)
		spr.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, cursor.time, cursor.time+700, 0.3, 0.0))
		spr.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, cursor.time, cursor.time+700, 0.05, 0.5))
		spr.ResetValuesToTransforms()
		spr.AdjustTimesToTransformations()
		spr.ShowForever(false)

		cursor.rippleContainer.Add(spr)
	}

	if cursor.lastLeftState != leftState || cursor.lastRightState != rightState {
		if leftState || rightState {
			cursor.scale.AddEventS(cursor.scale.GetTime(), cursor.scale.GetTime()+100, 1.0, 1.3)
		} else {
			cursor.scale.AddEventS(cursor.scale.GetTime(), cursor.scale.GetTime()+100, cursor.scale.GetValue(), 1.0)
		}

		cursor.lastLeftState = leftState
		cursor.lastRightState = rightState
	}

	cursor.smokeUpdate()

	cursor.scale.UpdateD(delta)

	cursor.renderer.Update(delta, cursor.Position)

	cursor.rippleContainer.Update(cursor.time)
}

func (cursor *Cursor) smokeUpdate() {
	if !settings.Cursor.SmokeEnabled || settings.PLAYERS != 1 {
		return
	}

	if cursor.SmokeKey && settings.PLAYERS == 1 {
		if !cursor.lastSmokeKey {
			cursor.lastSmokePosition = cursor.Position
			cursor.firstSmokePosition = cursor.Position
		}

		distance := math32.Max(2*scaling, cursor.smokeTexture.Width*scaling/4)
		points := cursor.Position.Dst(cursor.lastSmokePosition)

		if int(points/distance) > 0 {
			temp := cursor.lastSmokePosition
			for i := distance; i < points; i += distance {
				temp = cursor.Position.Sub(cursor.lastSmokePosition).Scl(i / points).Add(cursor.lastSmokePosition)

				smoke := sprite.NewSpriteSingle(cursor.smokeTexture, cursor.time*1000+float64(i), temp.Copy64(), bmath.Origin.Centre)
				smoke.SetAdditive(true)
				smoke.SetRotation(rand.Float64() * 2 * math.Pi)
				smoke.SetScale(0.5 / scaling)
				smoke.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, cursor.time, cursor.time+4000, 0.6, 0.0))
				smoke.ResetValuesToTransforms()
				smoke.AdjustTimesToTransformations()
				smoke.ShowForever(false)

				cursor.smokePointCount++

				cursor.smokeContainer.Add(smoke)
			}

			cursor.lastSmokePosition = temp

			if cursor.smokePointCount > 30 && cursor.Position.Dst(cursor.firstSmokePosition) < 10*scaling {
				cursor.smokeBrighten()
			}
		}
	} else if cursor.lastSmokeKey {
		cursor.smokeBrighten()
	}

	cursor.lastSmokeKey = cursor.SmokeKey

	cursor.smokeContainer.Update(cursor.time)
}

func (cursor *Cursor) smokeBrighten() {
	smokes := cursor.smokeContainer.GetProcessedSprites()

	delay := 0.0

	for _, s := range smokes {
		if (s.GetEndTime() - s.GetStartTime()) < 5000 {
			s.ClearTransformations()
			s.AddTransform(animation.NewSingleTransform(animation.Fade, easing.InQuad, cursor.time+delay, cursor.time+delay+8000, 1.0, 0.0))
			s.SetEndTime(cursor.time + delay + 8000)

			delay += 2.0
		}
	}

	cursor.smokePointCount = 0
}

func (cursor *Cursor) UpdateRenderer() {
	newSettings := settings.Skin.Cursor.UseSkinCursor

	if newSettings != cursor.lastSetting {
		cursor.lastSetting = newSettings
		if cursor.lastSetting {
			cursor.renderer = newOsuRenderer()
		} else {
			cursor.renderer = newDanserRenderer()
		}
	}

	cursor.renderer.UpdateRenderer()
}

func BeginCursorRender() {
	useAdditive = settings.Cursor.AdditiveBlending && (settings.PLAYERS > 1 || settings.DIVIDES > 1 || settings.TAG > 1) && !settings.Skin.Cursor.UseSkinCursor

	if useAdditive {
		cursorSpaceFbo.Bind()
		cursorSpaceFbo.ClearColor(0.0, 0.0, 0.0, 0.0)
	}

	blend.Push()
	blend.Enable()
	blend.SetFunctionSeparate(blend.SrcAlpha, blend.OneMinusSrcAlpha, blend.One, blend.OneMinusSrcAlpha)
}

func EndCursorRender() {
	if useAdditive {
		cursorSpaceFbo.Unbind()

		fboBatch.Begin()
		cursorSpaceFBOSprite.Draw(0, fboBatch)
		fboBatch.End()
	}

	blend.Pop()
}

func (cursor *Cursor) Draw(scale float64, batch *batch.QuadBatch, color color2.Color) {
	cursor.DrawM(scale, batch, color, color)
}

func (cursor *Cursor) DrawM(scale float64, batch *batch.QuadBatch, color color2.Color, colorGlow color2.Color) {
	if cursor.rippleContainer.GetNumProcessed() > 0 || cursor.smokeContainer.GetNumProcessed() > 0 {
		batch.Begin()
		batch.SetAdditive(false)
		batch.ResetTransform()
		batch.SetColor(1, 1, 1, 1)
		batch.SetScale(scaling*scaling, scaling*scaling)
		batch.SetSubScale(1, 1)

		cursor.smokeContainer.Draw(cursor.time, batch)
		cursor.rippleContainer.Draw(cursor.time, batch)

		batch.End()
	}

	if useAdditive {
		cursorFbo.Bind()
		cursorFbo.ClearColor(0.0, 0.0, 0.0, 0.0)
	}

	cursor.renderer.DrawM(scale, cursor.scale.GetValue(), batch, color, colorGlow)

	if useAdditive {
		cursorFbo.Unbind()

		fboBatch.Begin()

		blend.Push()
		blend.SetFunction(blend.SrcAlpha, blend.One)

		cursorFBOSprite.Draw(0, fboBatch)
		fboBatch.Flush()

		blend.Pop()

		fboBatch.End()
	}
}
