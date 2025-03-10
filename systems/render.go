package systems

import (
	"fmt"
	"image/color"
	"math"

	"github.com/bobbyhiddn/ecs-asteroids/components"
	"github.com/bobbyhiddn/ecs-asteroids/ecs"
	"github.com/bobbyhiddn/ecs-asteroids/game"
	"github.com/bobbyhiddn/ecs-asteroids/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type RenderSystem struct {
	world       *ecs.World
	screen      *ebiten.Image
	gameScreen  *game.Screen
	scoreSystem *ScoreSystem
}

func NewRenderSystem(world *ecs.World, screen *ebiten.Image) *RenderSystem {
	return &RenderSystem{
		world:       world,
		screen:      screen,
		gameScreen:  game.NewScreen(),
		scoreSystem: NewScoreSystem(world),
	}
}

func (s *RenderSystem) Update(dt float64) {
	// No update logic needed for rendering
}

func drawDottedCircle(screen *ebiten.Image, x, y, radius float64, c color.Color) {
	numSegments := 16 // Reduced for larger dots
	for i := 0; i < numSegments; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numSegments)
		nextAngle := float64(i+1) * 2 * math.Pi / float64(numSegments)

		// Only draw every other segment for dotted effect
		if i%2 == 0 {
			x1 := x + radius*math.Cos(angle)
			y1 := y + radius*math.Sin(angle)
			x2 := x + radius*math.Cos(nextAngle)
			y2 := y + radius*math.Sin(nextAngle)
			ebitenutil.DrawLine(screen, x1, y1, x2, y2, c)
		}
	}
}

func (s *RenderSystem) Draw(screen *ebiten.Image) {
	positions := s.world.Components["components.Position"]
	renderables := s.world.Components["components.Renderable"]
	rotations := s.world.Components["components.Rotation"]
	players := s.world.Components["components.Player"]
	explosions := s.world.Components["components.Explosion"]

	// Draw all renderable entities
	for id, renderableInterface := range renderables {
		renderable := renderableInterface.(components.Renderable)
		if !renderable.Visible {
			continue
		}

		position, ok := positions[id].(components.Position)
		if !ok {
			continue
		}

		rotation := float64(0)
		if r, ok := rotations[id].(components.Rotation); ok {
			rotation = r.Angle
		}

		switch renderable.Type {
		case components.RenderableTypeShip:
			isThrusting := false
			if player, ok := players[id].(components.Player); ok {
				isThrusting = player.IsThrusting
			}
			render.DrawShip(screen, position.X, position.Y, rotation, isThrusting)
		case components.RenderableTypeBullet:
			render.DrawBullet(screen, position.X, position.Y)
		case components.RenderableTypeAsteroid:
			render.DrawAsteroid(screen, position.X, position.Y, rotation, renderable.Scale)
		case components.RenderableTypeExplosion:
			if explosion, ok := explosions[id].(components.Explosion); ok {
				render.DrawExplosion(screen, position.X, position.Y, explosion)
			}
		}
	}

	// Draw UI
	for _, player := range players {
		if p, ok := player.(components.Player); ok {
			// Draw score
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", p.Score), 10, 10)
			// Draw lives
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Lives: %d", p.Lives), 10, 30)

			// If game is over, draw high scores
			if p.IsGameOver {
				s.drawGameOver(screen, p.Score)
			}

			// Draw fire button (red dotted circle)
			drawDottedCircle(screen, 100, float64(s.gameScreen.Height()-100), 80, color.RGBA{255, 0, 0, 255})
		}
	}
}

func (s *RenderSystem) drawGameOver(screen *ebiten.Image, currentScore int) {
	centerX := s.gameScreen.CenterX()
	startY := s.gameScreen.CenterY() - 100

	// Draw Game Over text
	gameOverText := "GAME OVER"
	ebitenutil.DebugPrintAt(screen, gameOverText, int(centerX)-30, int(startY))

	// Draw current score
	currentScoreText := fmt.Sprintf("Your Score: %d", currentScore)
	ebitenutil.DebugPrintAt(screen, currentScoreText, int(centerX)-40, int(startY)+30)

	// Draw high scores
	ebitenutil.DebugPrintAt(screen, "HIGH SCORES", int(centerX)-40, int(startY)+60)

	topScores := s.scoreSystem.GetTopScores()
	for i, score := range topScores {
		if i >= 5 { // Show only top 5 scores
			break
		}
		scoreText := fmt.Sprintf("%d. %d pts", i+1, score.Value)
		ebitenutil.DebugPrintAt(screen, scoreText, int(centerX)-40, int(startY)+90+i*20)
	}

	// Draw restart instruction
	ebitenutil.DebugPrintAt(screen, "Press SPACE to restart", int(centerX)-60, int(startY)+200)
}
