package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

// placing
// fix the indexing shit

const (
	GAMEXRES  = 900
	GAMEYRES  = 900
	GRIDW     = 16
	GRIDH     = 16
	GRID_SZ_X = GAMEXRES / (GRIDW - 1)
	GRID_SZ_Y = GAMEYRES / (GRIDH - 1)
)

// thats number of verts, so theres actually 1 less grid square

type GameState int

const (
	BETWEEN_WAVE GameState = iota
	IN_WAVE
)

type Context struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	atlas    []*sdl.Texture
	chunks   []*mix.Chunk

	grid []int

	placingTile int
}

var context Context = Context{}

func main() {
	rand.Seed(time.Now().UnixNano())

	initSDL()
	defer teardownSDL()

	initIMG()
	defer teardownIMG()

	loadTextures() // using proc

	context.grid = makeGrid() // using fn

	var mouseX, mouseY int32
	var lmbDown bool

	running := true
	tStart := time.Now().UnixNano()
	//tCurrentStart := float64(tStart) / 1000000000
	//var tLastStart float64
	for running {
		tStart = time.Now().UnixNano()
		//tLastStart = tCurrentStart
		//tCurrentStart = float64(tStart) / 1000000000

		//dt := tCurrentStart - tLastStart

		// handle input
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				fmt.Println(t)
				running = false
				break
			case *sdl.KeyboardEvent:
				if t.State == sdl.PRESSED {
					for i := range tileDefs {
						if sdl.Keycode(tileDefs[i].placeHotkey) == t.Keysym.Sym {
							context.placingTile = i
						}
					}
				}
			case *sdl.MouseButtonEvent:
				if t.Button == sdl.BUTTON_LEFT {
					if t.State == sdl.PRESSED {
						lmbDown = true
						context.grid[getClickedVertex(t.X, t.Y)] = context.placingTile
					} else {
						lmbDown = false
					}
				}
			case *sdl.MouseMotionEvent:
				mouseX = t.X
				mouseY = t.Y
				if lmbDown {
					context.grid[getClickedVertex(t.X, t.Y)] = context.placingTile
				}
			}

			// update stuff goes here

			// render loop
			context.renderer.Clear()
			context.renderer.SetDrawColor(0, 0, 0, 255)
			context.renderer.FillRect(&sdl.Rect{0, 0, GAMEXRES, GAMEYRES})
			for x := 0; x < GRIDW-1; x++ {
				for y := 0; y < GRIDH-1; y++ {
					//fmt.Print("Tile: ", x, y)
					toRect := &sdl.Rect{
						int32(x) * GRID_SZ_X,
						int32(y) * GRID_SZ_Y,
						GRID_SZ_X, GRID_SZ_Y,
					}
					marchOrder := []int{x + GRIDW*y, x + 1 + GRIDW*y, x + GRIDW*(y+1), x + 1 + GRIDW*(y+1)}
					sort.Slice(marchOrder, func(i, j int) bool {
						return tileDefs[context.grid[marchOrder[i]]].height < tileDefs[context.grid[marchOrder[j]]].height
					})

					// in order of height, run self marching squares for each one and draw
					for i, currentTile := range marchOrder {
						currentTileType := context.grid[currentTile]
						if i == 0 {
							context.renderer.CopyEx(tileDefs[currentTileType].texture, &sdl.Rect{0, 0, 64, 64}, toRect, 0, nil, sdl.FLIP_NONE)
							continue
						}
						marchResult := 0
						if context.grid[x+GRIDW*y] == currentTileType {
							marchResult += 8
						}
						if context.grid[x+1+GRIDW*y] == currentTileType {
							marchResult += 4
						}
						if context.grid[x+GRIDW*(y+1)] == currentTileType {
							marchResult += 2
						}
						if context.grid[x+1+GRIDW*(y+1)] == currentTileType {
							marchResult += 1
						}
						//fmt.Println("March result for", tileDefs[currentTileType].filename, ":", marchResult)

						// make srcRect from march result
						sx := 64 * (marchResult % 4)
						sy := 64 * (marchResult / 4)
						srcRect := &sdl.Rect{int32(sx), int32(sy), 64, 64}

						context.renderer.CopyEx(tileDefs[currentTileType].texture, srcRect, toRect, 0, nil, sdl.FLIP_NONE)
					}
				}
			}

			/*
				var i int32
				for i = 0; i < GRIDW; i++ {
					context.renderer.FillRect(&sdl.Rect{-1 + i*GAMEXRES/GRIDW, 0, 2, GAMEYRES})
				}
				for i = 0; i < GRIDH; i++ {
					context.renderer.FillRect(&sdl.Rect{0, -1 + i*GAMEYRES/GRIDH, GAMEXRES, 2})
				}
			*/
			cx, cy := getVertex(getClickedVertex(mouseX, mouseY))
			context.renderer.SetDrawColor(100, 200, 50, 128)
			context.renderer.FillRect(&sdl.Rect{cx - 32, cy - 32, 64, 64})

			context.renderer.SetDrawColor(200, 200, 200, 255)
			context.renderer.FillRect(&sdl.Rect{100, 700, 120, 120})
			context.renderer.CopyEx(
				tileDefs[context.placingTile].texture,
				&sdl.Rect{0, 0, 64, 64},
				&sdl.Rect{110, 710, 100, 100},
				0, nil, sdl.FLIP_NONE)

			context.renderer.Present()
			tnow := time.Now().UnixNano()
			currdt := tnow - tStart
			c := 1000000000/60 - currdt
			if c > 0 {
				time.Sleep(time.Nanosecond * time.Duration(c))
			}
		}
	}
}

func makeGrid() []int {
	grid := []int{}
	levelString := `
		dddddddddddddddd
		dddqdddddddddddd
		dmmmobbqdddddddd
		dmmobooooooddddd
		dttbtotqttdddddd
		dmmmobbqdddddddd
		dmmmbbbqdddddddd
		dddqdddddddddddd
		dddddddddddddddd
		dddddddddddddddd
		dddddddddddddddd
		dddddddddddddddd
		dddddddddddddddd
		dddddddddddddddd
		dddddddddddddddd
		dddddddddddddddd`

	for _, c := range levelString {
		for i := range tileDefs {
			if tileDefs[i].fileSym == c {
				grid = append(grid, i)
			}
		}
	}
	return grid
}

type TileDef struct {
	filename    string
	placeHotkey rune
	fileSym     rune
	height      int
	texture     *sdl.Texture
}

var tileDefs = [...]TileDef{
	TileDef{"City_BlackMarble", '1', 'm', 4, nil},
	TileDef{"City_BrickTiles", '2', 'b', 8, nil},
	TileDef{"City_Dirt", '3', 'd', 1, nil},
	TileDef{"City_DirtRough", '4', 'r', 2, nil},
	TileDef{"City_Grass", '5', 'g', 16, nil},
	TileDef{"City_GrassTrim", '6', 't', 15, nil},
	TileDef{"City_RoundTiles", '7', 'o', 12, nil},
	TileDef{"City_SquareTiles", '8', 'q', 10, nil},
	TileDef{"City_WhiteMarble", '9', 'w', 20, nil},
}

func loadTextures() {
	for i := range tileDefs {
		tileDefs[i].texture = loadTexture("assets/" + tileDefs[i].filename + ".png")
	}
}

func initSDL() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow("Autotile", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		GAMEXRES, GAMEYRES, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	context.window = window

	renderer, err := sdl.CreateRenderer(context.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	context.renderer = renderer
	context.renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
}

func initIMG() {
	fmt.Print("Init IMG PNG")
	if imgflags := img.Init(img.INIT_PNG); imgflags != img.INIT_PNG {
		panic("failed to init png loading")
	}
}

func teardownIMG() {
	fmt.Println("Quitting IMG")
	img.Quit()
}

func teardownSDL() {
	fmt.Print("Tearing down SDL...")
	for i := range tileDefs {
		tileDefs[i].texture.Destroy()
	}
	context.window.Destroy()
	context.renderer.Destroy()
	sdl.Quit()
	fmt.Println("Done")
}

func loadTexture(path string) *sdl.Texture {
	image, err := img.Load(path)
	if err != nil {
		panic(err)
	}
	defer image.Free()
	image.SetColorKey(true, 0xffff00ff)
	texture, err := context.renderer.CreateTextureFromSurface(image)
	if err != nil {
		panic(err)
	}
	texture.SetBlendMode(sdl.BLENDMODE_BLEND)
	return texture
}

func getClickedVertex(x, y int32) int {
	return int((x+GRID_SZ_X/2)/GRID_SZ_X + (y+GRID_SZ_Y/2)/GRID_SZ_Y*GRIDW)
}
func getVertex(i int) (int32, int32) {
	return int32((i % GRIDW) * GRID_SZ_X), int32((i / GRIDH) * GRID_SZ_Y)
}
