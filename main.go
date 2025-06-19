package main

import (
	"fmt"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

type Player struct {
	Pos          rl.Vector2
	Radius       float32
	Speed        float32
	Angle        float32
	LastFireTime float64 // Firing cooldown
}

type Bullet struct {
	Pos    rl.Vector2
	Vel    rl.Vector2
	Active bool
}

type Enemy struct {
	Pos    rl.Vector2
	Radius float32
	Speed  float32
	Active bool
}

type Wall struct {
	Rect rl.Rectangle
}

func main() {
	rl.InitWindow(screenWidth, screenHeight, "~CHOOTEMUP~")
	rl.SetTargetFPS(60)
	rl.DisableCursor()

	player := Player{
		Pos:    rl.NewVector2(screenWidth/2, screenHeight/2),
		Radius: 16,
		Speed:  300,
	}

	bullets := make([]Bullet, 50)
	for i := range bullets {
		bullets[i] = Bullet{Active: false}
	}

	enemies := make([]Enemy, 20)
	for i := range enemies {
		enemies[i] = Enemy{Active: false}
	}

	walls := []Wall{
		{Rect: rl.NewRectangle(0, 0, screenWidth, 32)},               // Top
		{Rect: rl.NewRectangle(0, screenHeight-32, screenWidth, 32)}, // Bottom
		{Rect: rl.NewRectangle(0, 0, 32, screenHeight)},              // Left
		{Rect: rl.NewRectangle(screenWidth-32, 0, 32, screenHeight)}, // Right
		{Rect: rl.NewRectangle(400, 200, 200, 32)},                   // Middle wall 1
		{Rect: rl.NewRectangle(600, 400, 32, 200)},                   // Middle wall 2
	}

	score := 0
	lives := 3
	gameOver := false

	for !rl.WindowShouldClose() {
		mousePos := rl.GetMousePosition()
		dx := mousePos.X - player.Pos.X
		dy := mousePos.Y - player.Pos.Y
		player.Angle = float32(math.Atan2(float64(dy), float64(dx)))

		// Handle input and movement
		if rl.IsKeyDown(rl.KeyW) {
			player.Pos.Y -= player.Speed * rl.GetFrameTime()
		}
		if rl.IsKeyDown(rl.KeyS) {
			player.Pos.Y += player.Speed * rl.GetFrameTime()
		}
		if rl.IsKeyDown(rl.KeyA) {
			player.Pos.X -= player.Speed * rl.GetFrameTime()
		}
		if rl.IsKeyDown(rl.KeyD) {
			player.Pos.X += player.Speed * rl.GetFrameTime()
		}

		// Player-wall collision check
		playerRect := rl.NewRectangle(player.Pos.X-player.Radius, player.Pos.Y-player.Radius, player.Radius*2, player.Radius*2)
		for _, wall := range walls {
			if rl.CheckCollisionRecs(playerRect, wall.Rect) {
				if rl.IsKeyDown(rl.KeyW) {
					player.Pos.Y += player.Speed * rl.GetFrameTime()
				}
				if rl.IsKeyDown(rl.KeyS) {
					player.Pos.Y -= player.Speed * rl.GetFrameTime()
				}
				if rl.IsKeyDown(rl.KeyA) {
					player.Pos.X += player.Speed * rl.GetFrameTime()
				}
				if rl.IsKeyDown(rl.KeyD) {
					player.Pos.X -= player.Speed * rl.GetFrameTime()
				}
			}
		}

		// Shooting
		if rl.IsKeyPressed(rl.KeyUp) {
			fireBullet(&bullets, player.Pos, rl.NewVector2(0, -1))
		}
		if rl.IsKeyPressed(rl.KeyDown) {
			fireBullet(&bullets, player.Pos, rl.NewVector2(0, 1))
		}
		if rl.IsKeyPressed(rl.KeyLeft) {
			fireBullet(&bullets, player.Pos, rl.NewVector2(-1, 0))
		}
		if rl.IsKeyPressed(rl.KeyRight) {
			fireBullet(&bullets, player.Pos, rl.NewVector2(1, 0))
		}

		// Mouse input handling
		const fireRate = 0.2
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) && rl.GetTime()-player.LastFireTime > fireRate {
			// Calculate bullet direction from player's angle
			dirX := float32(math.Cos(float64(player.Angle)))
			dirY := float32(math.Sin(float64(player.Angle)))
			dir := rl.NewVector2(dirX, dirY)

			fireBullet(&bullets, player.Pos, dir)
			player.LastFireTime = rl.GetTime()
		}

		// Update bullets
		for i := range bullets {
			if bullets[i].Active {
				bullets[i].Pos.X += bullets[i].Vel.X * 400 * rl.GetFrameTime()
				bullets[i].Pos.Y += bullets[i].Vel.Y * 400 * rl.GetFrameTime()
				if bullets[i].Pos.X < 0 || bullets[i].Pos.X > screenWidth || bullets[i].Pos.Y < 0 || bullets[i].Pos.Y > screenHeight {
					bullets[i].Active = false
				}
			}
		}

		// Spawn enemies
		if rand.Float32() < 0.02 {
			spawnEnemy(&enemies)
		}

		// Update Enemies
		for i := range enemies {
			if enemies[i].Active {
				dir := rl.Vector2Subtract(player.Pos, enemies[i].Pos)
				dir = rl.Vector2Normalize(dir)
				enemies[i].Pos.X += dir.X * enemies[i].Speed * rl.GetFrameTime()
				enemies[i].Pos.Y += dir.Y * enemies[i].Speed * rl.GetFrameTime()
			}
		}

		// Bullet-enemy collision
		for i := range bullets {
			if !bullets[i].Active {
				continue
			}
			for j := range enemies {
				if !enemies[j].Active {
					continue
				}
				if rl.CheckCollisionCircleRec(bullets[i].Pos, 4, rl.NewRectangle(enemies[j].Pos.X-enemies[j].Radius, enemies[j].Pos.Y-enemies[j].Radius, enemies[j].Radius*2, enemies[j].Radius*2)) {
					bullets[i].Active = false
					enemies[j].Active = false
					score += 10
					break
				}
			}
		}

		// Player-enemy collision
		for _, enemy := range enemies {
			if enemy.Active && rl.CheckCollisionCircles(player.Pos, player.Radius, enemy.Pos, enemy.Radius) {
				lives--
				if lives > 0 {
					player.Pos = rl.NewVector2(screenWidth/2, screenHeight/2)
				} else {
					gameOver = true
				}
				break // Prevent multiple hits in one frame
			}
		}

		// Draw
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		for _, wall := range walls {
			rl.DrawRectangleRec(wall.Rect, rl.DarkGray)
		}
		rl.DrawCircleV(player.Pos, player.Radius, rl.DarkGreen)
		directionLength := float32(30)
		endX := player.Pos.X + float32(math.Cos(float64(player.Angle)))*directionLength
		endY := player.Pos.Y + float32(math.Sin(float64(player.Angle)))*directionLength
		rl.SetLineWidth(4.0)
		rl.DrawLineV(player.Pos, rl.NewVector2(endX, endY), rl.DarkGreen)

		for _, bullet := range bullets {
			if bullet.Active {
				rl.DrawRectangle(int32(bullet.Pos.X-4), int32(bullet.Pos.Y-4), 8, 8, rl.Yellow)
			}
		}
		for _, enemy := range enemies {
			if enemy.Active {
				rl.DrawCircleV(enemy.Pos, enemy.Radius, rl.Red)
			}
		}

		if gameOver {
			rl.DrawText("GAME OVER! Press R to Restart", screenWidth/2-200, screenHeight/2, 30, rl.Maroon)
		} else {
			rl.DrawText(fmt.Sprintf("Score: %d", score), 10, 10, 20, rl.Gold)
			rl.DrawText(fmt.Sprintf("Lives: %d", lives), 10, 40, 20, rl.Green)
		}

		rl.EndDrawing()

		if gameOver && rl.IsKeyPressed(rl.KeyR) {
			lives = 3
			score = 0
			player.Pos = rl.NewVector2(screenWidth/2, screenHeight/2)
			for i := range enemies {
				enemies[i].Active = false // Reset enemies
			}
			gameOver = false
		}
	}

	rl.CloseWindow()
}

func fireBullet(bullets *[]Bullet, pos rl.Vector2, dir rl.Vector2) {
	for i := range *bullets {
		if !(*bullets)[i].Active {
			(*bullets)[i] = Bullet{
				Pos:    pos,
				Vel:    dir,
				Active: true,
			}
			break
		}
	}
}

func spawnEnemy(enemies *[]Enemy) {
	for i := range *enemies {
		if !(*enemies)[i].Active {
			var pos rl.Vector2
			side := rand.Intn(4)
			switch side {
			case 0: // Top
				pos = rl.NewVector2(float32(rand.Intn(screenWidth)), 0)
			case 1: // Bottom
				pos = rl.NewVector2(float32(rand.Intn(screenWidth)), screenHeight)
			case 2: // Left
				pos = rl.NewVector2(0, float32(rand.Intn(screenHeight)))
			case 3: // Right
				pos = rl.NewVector2(screenWidth, float32(rand.Intn(screenHeight)))
			}
			(*enemies)[i] = Enemy{
				Pos:    pos,
				Radius: 16,
				Speed:  80,
				Active: true,
			}
			break
		}
	}
}
