package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth      = 800
	screenHeight     = 800
	cellSize         = 10
	initialCells     = 2000
	fadeAlpha        = 0x10
	zoomStep         = 1.05
	minZoom, maxZoom = 0.25, 5.0
)

// Padded world to prevent black bars when zoomed out
var (
	worldWidth  = 3000 + screenWidth
	worldHeight = 3000 + screenHeight
)

type Cell struct {
	X, Y int
}

type Game struct {
	cells      []Cell
	trailImage *ebiten.Image
	frameCount int
	zoom       float64
	cameraX    float64
	cameraY    float64
}

func NewGame() *Game {
	cells := make([]Cell, 0, initialCells)
	occupied := map[[2]int]bool{}
	for len(cells) < initialCells {
		x := rand.Intn(worldWidth / cellSize)
		y := rand.Intn(worldHeight / cellSize)
		pos := [2]int{x, y}
		if !occupied[pos] {
			occupied[pos] = true
			cells = append(cells, Cell{X: x, Y: y})
		}
	}

	trail := ebiten.NewImage(worldWidth, worldHeight)
	trail.Fill(color.Black)

	return &Game{
		cells:      cells,
		trailImage: trail,
		zoom:       1.0,
		cameraX:    float64(worldWidth) / 2,
		cameraY:    float64(worldHeight) / 2,
	}
}

func (g *Game) screenToWorld(screenX, screenY float64) (float64, float64) {
	cx, cy := float64(screenWidth)/2, float64(screenHeight)/2
	wx := (screenX - cx) / g.zoom + g.cameraX
	wy := (screenY - cy) / g.zoom + g.cameraY
	return wx, wy
}

func (g *Game) clampCamera() {
	halfW := float64(screenWidth) / 2 / g.zoom
	halfH := float64(screenHeight) / 2 / g.zoom

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

	if g.cameraX < minX {
		g.cameraX = minX
	}
	if g.cameraX > maxX {
		g.cameraX = maxX
	}
	if g.cameraY < minY {
		g.cameraY = minY
	}
	if g.cameraY > maxY {
		g.cameraY = maxY
	}
}

func (g *Game) Update() error {
	g.frameCount++

	mouseX, mouseY := ebiten.CursorPosition()
	mx, my := float64(mouseX), float64(mouseY)

	beforeX, beforeY := g.screenToWorld(mx, my)

	// Zoom via Q/E keys
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.zoom *= zoomStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.zoom /= zoomStep
	}

	// Zoom via mouse wheel
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		if scrollY > 0 {
			g.zoom *= zoomStep
		} else if scrollY < 0 {
			g.zoom /= zoomStep
		}
	}

	// Clamp zoom range
	if g.zoom < minZoom {
		g.zoom = minZoom
	}
	if g.zoom > maxZoom {
		g.zoom = maxZoom
	}

	afterX, afterY := g.screenToWorld(mx, my)
	g.cameraX += beforeX - afterX
	g.cameraY += beforeY - afterY

	// WASD camera movement
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

	g.clampCamera()

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
		if !newPositions[newPos] {
			newPositions[newPos] = true
			newCells = append(newCells, Cell{X: nx, Y: ny})
		}
	}

	g.cells = newCells
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Fade trail
	fadeRect := ebiten.NewImage(worldWidth, worldHeight)
	fadeRect.Fill(color.RGBA{0, 0, 0, fadeAlpha})
	g.trailImage.DrawImage(fadeRect, &ebiten.DrawImageOptions{})

	// Draw cells
	for _, cell := range g.cells {
		x := float64(cell.X * cellSize)
		y := float64(cell.Y * cellSize)
		ebitenutil.DrawRect(g.trailImage, x, y, cellSize, cellSize, color.White)
	}

	op := &ebiten.DrawImageOptions{}
	cx, cy := float64(screenWidth)/2, float64(screenHeight)/2
	op.GeoM.Translate(-g.cameraX, -g.cameraY)
	op.GeoM.Scale(g.zoom, g.zoom)
	op.GeoM.Translate(cx, cy)

	screen.DrawImage(g.trailImage, op)

	info := fmt.Sprintf("FPS: %.2f  Zoom: %.2fx  Cells: %d", ebiten.CurrentTPS(), g.zoom, len(g.cells))
	ebitenutil.DebugPrintAt(screen, info, 10, 10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Zoomy Zombies (Now With Mouse Wheel)")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
