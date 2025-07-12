package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	initialWidth      = 800
	initialHeight     = 800
	cellSize          = 10
	zoomStep          = 1.05
	minZoom, maxZoom  = 0.25, 5.0
	fullscreenKeyCode = ebiten.KeyF11
	minSpeed          = 0.1
	maxSpeed          = 10.0
	speedStep = 0.1
)

var (
	worldWidth  = 4000
	worldHeight = 4000
)

type LifeType string

const (
	Life   LifeType = "life"
	Zombie LifeType = "zombie"
)

type Cell struct {
	X, Y int
	Type LifeType
}

type Game struct {
	cells        []Cell
	zoom         float64
	cameraX      float64
	cameraY      float64
	fullscreen   bool
	prevF11Down  bool
	gameSpeed    float64
	speedCounter float64
	occupied     map[[2]int]bool
	placeType    LifeType
	showUI       bool
}

func NewGame() *Game {
	return &Game{
		cells:     []Cell{},
		zoom:      1.0,
		cameraX:   float64(worldWidth) / 2,
		cameraY:   float64(worldHeight) / 2,
		gameSpeed: 1.0,
		occupied:  make(map[[2]int]bool),
		placeType: Life,
		showUI:    true,
	}
}

func (g *Game) screenToWorld(screenX, screenY float64, screenW, screenH float64) (float64, float64) {
	cx, cy := screenW/2, screenH/2
	wx := (screenX - cx)/g.zoom + g.cameraX
	wy := (screenY - cy)/g.zoom + g.cameraY
	return wx, wy
}

func (g *Game) clampCamera(screenW, screenH float64) {
	halfW := screenW / 2 / g.zoom
	halfH := screenH / 2 / g.zoom

	minX := halfW
	minY := halfH
	maxX := float64(worldWidth) - halfW
	maxY := float64(worldHeight) - halfH

	if maxX < minX {
		maxX = minX
	}
	if maxY < minY {
		maxY = minY
	}

	g.cameraX = math.Max(minX, math.Min(g.cameraX, maxX))
	g.cameraY = math.Max(minY, math.Min(g.cameraY, maxY))
}

func (g *Game) Update() error {
	screenW, screenH := ebiten.WindowSize()
	mouseX, mouseY := ebiten.CursorPosition()
	mx, my := float64(mouseX), float64(mouseY)
	beforeX, beforeY := g.screenToWorld(mx, my, float64(screenW), float64(screenH))

	f11Down := ebiten.IsKeyPressed(fullscreenKeyCode)
	if f11Down && !g.prevF11Down {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}
	g.prevF11Down = f11Down

	if ebiten.IsKeyPressed(ebiten.KeyTab) {
		g.showUI = true
	} else {
		g.showUI = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.gameSpeed = math.Min(g.gameSpeed+speedStep, maxSpeed)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.gameSpeed = math.Max(g.gameSpeed-speedStep, minSpeed)
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.zoom *= zoomStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.zoom /= zoomStep
	}
	_, scrollY := ebiten.Wheel()
	if scrollY > 0 {
		g.zoom *= zoomStep
	} else if scrollY < 0 {
		g.zoom /= zoomStep
	}
	g.zoom = math.Max(minZoom, math.Min(g.zoom, maxZoom))

	afterX, afterY := g.screenToWorld(mx, my, float64(screenW), float64(screenH))
	g.cameraX += beforeX - afterX
	g.cameraY += beforeY - afterY

	moveSpeed := 10.0 / g.zoom
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.cameraY -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.cameraY += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.cameraX -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.cameraX += moveSpeed
	}

	if ebiten.IsKeyPressed(ebiten.Key1) {
		g.placeType = Life
	}
	if ebiten.IsKeyPressed(ebiten.Key2) {
		g.placeType = Zombie
	}

	g.clampCamera(float64(screenW), float64(screenH))

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		wx, wy := g.screenToWorld(mx, my, float64(screenW), float64(screenH))
		cx := int(wx) / cellSize
		cy := int(wy) / cellSize
		pos := [2]int{cx, cy}
		if !g.occupied[pos] && cx >= 0 && cy >= 0 && cx*cellSize < worldWidth && cy*cellSize < worldHeight {
			g.occupied[pos] = true
			g.cells = append(g.cells, Cell{X: cx, Y: cy, Type: g.placeType})
		}
	}

	g.speedCounter += g.gameSpeed
	if g.speedCounter >= 1.0 {
		updates := int(math.Floor(g.speedCounter))
		for i := 0; i < updates; i++ {
			g.logicUpdate()
		}
		g.speedCounter -= float64(updates)
	}

	return nil
}

func (g *Game) logicUpdate() {
	newPositions := map[[2]int]bool{}
	newCells := make([]Cell, 0, len(g.cells))

	for _, cell := range g.cells {
		dx := rand.Intn(3) - 1
		dy := rand.Intn(3) - 1
		nx := cell.X + dx
		ny := cell.Y + dy

		if nx < 0 || ny < 0 || nx*cellSize >= worldWidth || ny*cellSize >= worldHeight {
			nx, ny = cell.X, cell.Y
		}

		newPos := [2]int{nx, ny}
		origPos := [2]int{cell.X, cell.Y}

		if !newPositions[newPos] {
			newPositions[newPos] = true
			newCells = append(newCells, Cell{X: nx, Y: ny, Type: cell.Type})
		} else if !newPositions[origPos] {
			newPositions[origPos] = true
			newCells = append(newCells, cell)
		} else {
			newCells = append(newCells, cell)
			newPositions[origPos] = true
		}
	}

	g.occupied = newPositions
	g.cells = newCells
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	screenW, screenH := ebiten.WindowSize()
	cx, cy := float64(screenW)/2, float64(screenH)/2

	lifeCount := 0
	zombieCount := 0

	for _, cell := range g.cells {
		x := float64(cell.X * cellSize)
		y := float64(cell.Y * cellSize)
		screenX := (x - g.cameraX)*g.zoom + cx
		screenY := (y - g.cameraY)*g.zoom + cy

		if screenX >= 0 && screenX < float64(screenW) && screenY >= 0 && screenY < float64(screenH) {
			var col color.Color
			switch cell.Type {
			case Life:
				col = color.White
				lifeCount++
			case Zombie:
				col = color.RGBA{0, 255, 0, 255}
				zombieCount++
			}

			// Scale and center cell size based on zoom
			size := math.Max(1.0, float64(cellSize)*g.zoom)
			offset := (size - float64(cellSize)) / 2
			ebitenutil.DrawRect(screen, screenX-offset, screenY-offset, size, size, col)
		}
	}

	info := fmt.Sprintf("FPS: %.2f  Zoom: %.2fx  Speed: %.2fx  Total: %d  Life: %d  Zombie: %d  [1:Life 2:Zombie]  Current: %s",
		ebiten.CurrentTPS(), g.zoom, g.gameSpeed, len(g.cells), lifeCount, zombieCount, g.placeType)
	ebitenutil.DebugPrintAt(screen, info, 10, 10)

	if g.showUI {
		uiColor := color.RGBA{50, 50, 50, 200}
		panelWidth := 250
		panelHeight := 160
		ebitenutil.DrawRect(screen, 10, 30, float64(panelWidth), float64(panelHeight), uiColor)
		ebitenutil.DebugPrintAt(screen, "[TAB] Toggle UI", 20, 40)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[Q/E or Scroll] Zoom: %.2fx", g.zoom), 20, 60)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[←/→] Speed: %.1fx", g.gameSpeed), 20, 80)
		ebitenutil.DebugPrintAt(screen, "[1] Place Life  [2] Place Zombie", 20, 100)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Current: %s", g.placeType), 20, 120)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Life: %d  Zombie: %d", lifeCount, zombieCount), 20, 140)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetWindowSize(initialWidth, initialHeight)
	ebiten.SetWindowTitle("Lifes Sandbox — F11 for Fullscreen")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
