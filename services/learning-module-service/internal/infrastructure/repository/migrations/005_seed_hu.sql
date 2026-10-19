UPDATE courses SET 
    title_hu = 'Tanulj Pythont', 
    description_hu = 'Üdvözöllek a "Tanulj Pythont" kurzuson! Ebből a kurzusból megtanulhatsz Pythonban programozni, amely a világ egyik legnépszerűbb és legsokoldalúbb programozási nyelve.' 
WHERE id = 'python-basics';

UPDATE chapters SET title_hu = '1. Fejezet: Változók' WHERE id = 'ch1-variables';

UPDATE lessons SET 
    title_hu = '1. Lecke: A Python alapjai', 
    description_hu = E'Ebben a kurzusban egy képzeletbeli szoftverprojekt részeit fogjuk felépíteni: a Textio-t! Egy SaaS alkalmazást, amely SMS-eket küld ügyfeleknek.\n\nSzerkeszd a kódot úgy, hogy az `edit me` helyett a `starting server...` szöveget írja ki a konzolra.' 
WHERE id = 'l1-intro';

UPDATE lessons SET 
    title_hu = '2. Lecke: Változók', 
    description_hu = 'Pythonban a változókat egyszerűen úgy hozod létre, hogy értéket adsz egy névnek. Az alábbiak közül melyik érvényes változónév?' 
WHERE id = 'l2-variables';

UPDATE lessons SET 
    title_hu = '3. Lecke: Sztringek', 
    description_hu = 'A sztringek karaktersorozatok a Pythonban. Hozz létre egy `greeting` nevű változót, amely a `"Hello, World!"` értéket tartalmazza.' 
WHERE id = 'l3-string';

UPDATE answer_options SET text_hu = '1st_valtozo' WHERE id = 'opt1';
UPDATE answer_options SET text_hu = 'my_var' WHERE id = 'opt2';
UPDATE answer_options SET text_hu = 'my-var' WHERE id = 'opt3';
UPDATE answer_options SET text_hu = 'osztaly' WHERE id = 'opt4';

UPDATE chapters SET title_hu = '2. Fejezet: Függvények és Logika' WHERE id = 'ch2-functions';

UPDATE lessons SET 
    title_hu = '1. Lecke: Ciklusok', 
    description_hu = E'A ciklusok (loops) segítségével egy utasítássort többször is végrehajthatunk, amíg egy adott feltétel teljesül. A Pythonban két fő ciklusfajta létezik: a `for` és a `while` ciklus.\n\nPéldául egy lista elemein így iterálhatunk végig:\n\n```python\ngyumolcsok = ["alma", "banán", "cseresznye"]\nfor x in gyumolcsok:\n  print(x)\n```'
WHERE id = 'l4-theory-loops';

UPDATE lessons SET 
    title_hu = '3. Lecke: Függvények definiálása', 
    description_hu = 'Melyik kulcsszót használjuk a Pythonban egy új függvény definiálására?' 
WHERE id = 'l5-quiz-def';

UPDATE lessons SET 
    title_hu = '4. Lecke: Faktoriális', 
    description_hu = E'Írj egy `factorial(n)` nevű függvényt, amely kiszámítja a paraméterként kapott `n` egész szám faktoriálisát!\n\nA faktoriális (jelölése n!) az 1-től n-ig terjedő egész számok szorzata. Például `factorial(5)` eredménye `120` (1 * 2 * 3 * 4 * 5). A 0 faktoriálisa 1.' 
WHERE id = 'l6-code-factorial';

UPDATE lessons SET 
    title_hu = '5. Lecke: Szótárak (Dictionary)', 
    description_hu = E'A Python `dictionary` (szótár) kulcs-érték párokat tárol.\n\nEgészítsd ki a hiányzó részeket (`___`), hogy a program a "kor" értékét (30) írja ki a konzolra!' 
WHERE id = 'l7-comp-dict';

UPDATE answer_options SET text_hu = 'func' WHERE id = 'opt-py-new-1';
UPDATE answer_options SET text_hu = 'def' WHERE id = 'opt-py-new-2';
UPDATE answer_options SET text_hu = 'function' WHERE id = 'opt-py-new-3';
UPDATE answer_options SET text_hu = 'define' WHERE id = 'opt-py-new-4';
