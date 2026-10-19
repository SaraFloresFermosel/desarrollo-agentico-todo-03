UPDATE tasks SET position = t.rn
FROM (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC, id ASC) AS rn
  FROM tasks
) AS t
WHERE tasks.id = t.id;
