-- Shift orders of existing lessons
UPDATE lessons SET order_num = 5 WHERE id = 'l7-comp-dict';
UPDATE lessons SET order_num = 4 WHERE id = 'l6-code-factorial';
UPDATE lessons SET order_num = 3 WHERE id = 'l5-quiz-def';

-- Insert new theory lesson for functions
INSERT INTO lessons (id, chapter_id, order_num, title, description, type, language, starter_code, solution_code, test_code, expected_output) VALUES
(
    'l4b-theory-functions',
    'ch2-functions',
    2,
    'L5: Functions',
    'A function is a block of code which only runs when it is called. You can pass data, known as parameters, into a function. A function can return data as a result.',
    'theory',
    NULL,
    NULL,
    NULL,
    NULL,
    NULL
) ON CONFLICT (id) DO NOTHING;

-- Update English loop content
UPDATE lessons SET description = E'Loops can execute a block of code a number of times.\n\nPython has two primitive loop commands:\n* `while` loops\n* `for` loops\n\n```python\ni = 1\nwhile i < 6:\n  print(i)\n  i += 1\n```' WHERE id = 'l4-theory-loops';

-- Update Hungarian loop content
UPDATE lessons SET description_hu = E'A ciklusok segítségével egy kódblokkot többször is végrehajthatunk.\n\nA Pythonban két fő ciklusfajta létezik:\n* **`while` ciklus**: addig fut, amíg egy feltétel igaz.\n* **`for` ciklus**: egy sorozaton (például listán) iterál végig.\n\n### Példa egy for ciklusra:\n\n```python\ngyumolcsok = ["alma", "banán", "cseresznye"]\nfor x in gyumolcsok:\n  print(x)\n```\n\n### Példa egy while ciklusra:\n\n```python\ni = 1\nwhile i < 6:\n  print(i)\n  i += 1\n```' WHERE id = 'l4-theory-loops';

-- Add Hungarian translations for the new function theory
UPDATE lessons SET 
    title_hu = '2. Lecke: Függvények', 
    description_hu = E'A függvény (function) egy olyan kódblokk, amely csak akkor fut le, ha meghívják.\n\nA függvényeknek adhatunk át adatokat, amelyeket **paramétereknek** nevezünk. A függvény eredményként is visszaadhat adatokat a `return` kulcsszó segítségével.\n\nPythonban egy függvényt a `def` kulcsszóval definiálunk:\n\n```python\ndef my_function():\n  print("Hello a függvényből!")\n\n# Függvény meghívása:\nmy_function()\n```\n\nParaméteres függvény:\n\n```python\ndef greet(name):\n  print("Hello, " + name)\n\ngreet("Anna")\n```' 
WHERE id = 'l4b-theory-functions';
