-- name: GetOutfit :one
SELECT * FROM outfits
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListOutfits :many
SELECT * FROM outfits
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateOutfit :one
INSERT INTO outfits (
  user_id, name, notes, wear_count, rating, body_image_url
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateOutfit :one
UPDATE outfits
SET name = $2,
    notes = $3,
    rating = $4,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteOutfit :exec
DELETE FROM outfits
WHERE id = $1 AND user_id = $2;

-- name: IncrementOutfitWearCount :exec
UPDATE outfits
SET wear_count = wear_count + 1
WHERE id = $1;

-- name: AddItemToOutfit :one
INSERT INTO outfit_items (
  outfit_id, item_id, position_x, position_y, position_z, scale
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetOutfitItems :many
SELECT oi.*, i.*
FROM outfit_items oi
JOIN items i ON i.id = oi.item_id
WHERE oi.outfit_id = $1
ORDER BY oi.position_z ASC;

-- name: DeleteOutfitItems :exec
DELETE FROM outfit_items
WHERE outfit_id = $1;
