-- 014_daily_challenges.sql: Daily Challenge tables and seed data

CREATE TABLE IF NOT EXISTS daily_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    explanation TEXT NOT NULL,
    difficulty VARCHAR(20) DEFAULT 'Közepes',
    points_reward INT DEFAULT 40,
    xp_reward INT DEFAULT 80,
    starter_codes JSONB NOT NULL,
    test_cases JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_daily_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id UUID NOT NULL REFERENCES daily_challenges(id) ON DELETE CASCADE,
    challenge_date DATE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    solve_time_seconds INT,
    language VARCHAR(50) NOT NULL,
    submitted_code TEXT,
    is_completed BOOLEAN DEFAULT false,
    points_earned INT DEFAULT 0,
    xp_earned INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, challenge_date)
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS challenge_streak INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_challenge_date DATE;

-- Seed Challenges
INSERT INTO daily_challenges (slug, title, description, explanation, difficulty, points_reward, xp_reward, starter_codes, test_cases)
VALUES 
(
    'two-sum',
    'Két szám összege (Two Sum)',
    'Adott egy egész számokból álló lista (`nums`) és egy célszám (`target`). Keresd meg azt a két különböző indexet, amelyekhez tartozó számok összege pontosan a célszám!',
    'A feladatban egy számlistát és egy célösszeget kapsz. Keresd meg a két olyan elem indexét, amelyek összege a cél.\n\n**Példa lépésről lépésre:**\nHa a lista `[2, 7, 11, 15]` és a cél `9`, akkor a `2` (0. index) és a `7` (1. index) összege `9`, tehát a válasz: `[0, 1]`.\n\n**Hasznos tipp:**\nEgy szótárban (Map / Dictionary) eltárolhatod a már látott számokat és azok indexeit, így gyorsan ellenőrizheted, hogy a szükséges különbség (`target - num`) szerepel-e már.',
    'Könnyű',
    40,
    80,
    '{
        "python": "def two_sum(nums, target):\n    # Írd ide a megoldásodat!\n    pass\n",
        "javascript": "function twoSum(nums, target) {\n    // Írd ide a megoldásodat!\n}\n",
        "csharp": "using System;\nusing System.Collections.Generic;\n\npublic class Solution {\n    public static int[] TwoSum(int[] nums, int target) {\n        // Írd ide a megoldásodat!\n        return new int[0];\n    }\n}\n",
        "php": "<?php\nfunction twoSum($nums, $target) {\n    // Írd ide a megoldásodat!\n}\n",
        "go": "package main\n\nfunc TwoSum(nums []int, target int) []int {\n    // Írd ide a megoldásodat!\n    return nil\n}\n"
    }'::jsonb,
    '[
        {"input": "[2, 7, 11, 15], 9", "expected": "[0, 1]"},
        {"input": "[3, 2, 4], 6", "expected": "[1, 2]"},
        {"input": "[3, 3], 6", "expected": "[0, 1]"}
    ]'::jsonb
),
(
    'palindrome-check',
    'Palindrom Szöveg Ellenőrző',
    'Határozd meg, hogy egy adott szöveg (`s`) palindrom-e (visszafelé olvasva ugyanaz-e), ha figyelmen kívül hagyjuk a szóközöket, írásjeleket és a kis/nagybetűk különbségeit!',
    'A palindrom egy olyan szöveg, ami előrefelé és visszafelé olvasva is betűre megegyezik.\n\n**Megoldás menete:**\n1. Alakíts minden betűt kisbetűssé.\n2. Távolítsd el a szóközöket és írásjeleket (csak betűk és számok maradjanak).\n3. Ellenőrizd, hogy az így kapott szöveg megegyezik-e a megfordított változatával.',
    'Könnyű',
    35,
    70,
    '{
        "python": "def is_palindrome(s):\n    # Írd ide a megoldásodat!\n    pass\n",
        "javascript": "function isPalindrome(s) {\n    // Írd ide a megoldásodat!\n}\n",
        "csharp": "using System;\nusing System.Text.RegularExpressions;\n\npublic class Solution {\n    public static bool IsPalindrome(string s) {\n        // Írd ide a megoldásodat!\n        return false;\n    }\n}\n",
        "php": "<?php\nfunction isPalindrome($s) {\n    // Írd ide a megoldásodat!\n}\n",
        "go": "package main\n\nfunc IsPalindrome(s string) bool {\n    // Írd ide a megoldásodat!\n    return false\n}\n"
    }'::jsonb,
    '[
        {"input": "\"A man, a plan, a canal: Panama\"", "expected": "true"},
        {"input": "\"race a car\"", "expected": "false"},
        {"input": "\"Gorog\"", "expected": "true"}
    ]'::jsonb
),
(
    'balanced-parentheses',
    'Zárójel Validátor',
    'Egy string csak a `(`, `)`, `{`, `}`, `[`, `]` karaktereket tartalmazza. Döntsd el, hogy a zárójelek párosítása szabályos-e (minden nyitó zárójelnek van megfelelő típusú és helyes sorrendű zárópárja)!',
    'A zárójelek akkor szabályosak, ha minden nyitott zárójel a megfelelő típusú zárópárral van lezárva, és helyes beágyazási sorrendben követik egymást.\n\n**Megoldási javaslat:**\nHasználj egy listát veremként (Stack):\n- Nyitó zárójel (`(`, `{`, `[`) esetén fűzd hozzá a verem végéhez.\n- Záró pár esetén ellenőrizd, hogy a verem legfelső eleme a hozzá illő nyitópár-e. Ha igen, vedd ki a veremből, ha nem, a zárójelezés hibás.',
    'Közepes',
    45,
    90,
    '{
        "python": "def is_valid(s):\n    # Írd ide a megoldásodat!\n    pass\n",
        "javascript": "function isValid(s) {\n    // Írd ide a megoldásodat!\n}\n",
        "csharp": "using System;\nusing System.Collections.Generic;\n\npublic class Solution {\n    public static bool IsValid(string s) {\n        // Írd ide a megoldásodat!\n        return false;\n    }\n}\n",
        "php": "<?php\nfunction isValid($s) {\n    // Írd ide a megoldásodat!\n}\n",
        "go": "package main\n\nfunc IsValid(s string) bool {\n    // Írd ide a megoldásodat!\n    return false\n}\n"
    }'::jsonb,
    '[
        {"input": "\"()[]{}\"", "expected": "true"},
        {"input": "\"(]\"", "expected": "false"},
        {"input": "\"([{}])\"", "expected": "true"},
        {"input": "\"([)]\"", "expected": "false"}
    ]'::jsonb
),
(
    'majority-element',
    'Többségi Elem Keresése',
    'Adott egy `n` méretű lista (`nums`). Keresd meg a többségi elemet, vagyis azt az értéket, amely több mint `n / 2` alkalommal fordul elő a listában!',
    'A többségi elem az a szám, amely a lista elemeinek több mint felét teszi ki.\n\n**Megoldási javaslat:**\nSzámold meg az elemek előfordulását egy szótárban vagy Map-ben. Amint valamelyik elem darabszáma meghaladja a lista felét (`len(nums) // 2`), azonnal megvan a válasz.',
    'Könnyű',
    40,
    80,
    '{
        "python": "def majority_element(nums):\n    # Írd ide a megoldásodat!\n    pass\n",
        "javascript": "function majorityElement(nums) {\n    // Írd ide a megoldásodat!\n}\n",
        "csharp": "using System;\nusing System.Collections.Generic;\n\npublic class Solution {\n    public static int MajorityElement(int[] nums) {\n        // Írd ide a megoldásodat!\n        return 0;\n    }\n}\n",
        "php": "<?php\nfunction majorityElement($nums) {\n    // Írd ide a megoldásodat!\n}\n",
        "go": "package main\n\nfunc MajorityElement(nums []int) int {\n    // Írd ide a megoldásodat!\n    return 0\n}\n"
    }'::jsonb,
    '[
        {"input": "[3, 2, 3]", "expected": "3"},
        {"input": "[2, 2, 1, 1, 1, 2, 2]", "expected": "2"},
        {"input": "[6, 5, 5]", "expected": "5"}
    ]'::jsonb
),
(
    'roman-to-integer',
    'Római Szám Átváltó',
    'Válts át egy római számot (pl. `"XIV"`, `"MCMXCIV"`) egész számmá! Értékek: I=1, V=5, X=10, L=50, C=100, D=500, M=1000.',
    '💡 **Magyarázat & Megoldási Tippek:**\n\n- Járd be a szöveget balról jobbra.\n- Ha egy karakter értéke **kisebb**, mint az utána következőé (pl. `IV` -> 1 < 5, vagy `IX` -> 1 < 10), akkor **vond le** az értékét az összegből.\n- Minden más esetben **add hozzá** az értékét.',
    'Közepes',
    45,
    90,
    '{
        "python": "def roman_to_int(s):\n    # Írd ide a megoldásodat!\n    pass\n",
        "javascript": "function romanToInt(s) {\n    // Írd ide a megoldásodat!\n}\n",
        "csharp": "using System;\nusing System.Collections.Generic;\n\npublic class Solution {\n    public static int RomanToInt(string s) {\n        // Írd ide a megoldásodat!\n        return 0;\n    }\n}\n",
        "php": "<?php\nfunction romanToInt($s) {\n    // Írd ide a megoldásodat!\n}\n",
        "go": "package main\n\nfunc RomanToInt(s string) int {\n    // Írd ide a megoldásodat!\n    return 0\n}\n"
    }'::jsonb,
    '[
        {"input": "\"III\"", "expected": "3"},
        {"input": "\"LVIII\"", "expected": "58"},
        {"input": "\"MCMXCIV\"", "expected": "1994"}
    ]'::jsonb
),
(
    'dragon-fizzbuzz',
    'Sárkány FizzBuzz RPG',
    'Készíts egy függvényt, amely 1-től `n`-ig egy listát ad vissza: ha a szám 3-mal osztható, legyen `"Sárkány"`, ha 5-tel, akkor `"Mágia"`, ha mindkettővel (15-tel), akkor `"Sárkánymágia"`, egyébként maga a szám szövegként!',
    '💡 **Magyarázat & Megoldási Tippek:**\n\n- Fontos a feltételek sorrendje! Mindig a 15-ös (3 és 5 egyszerre) oszthatóságot ellenőrizd először, mert a 15-re a 3-as és 5-ös feltétel is külön-külön igaz lenne.',
    'Könnyű',
    30,
    60,
    '{
        "python": "def dragon_fizzbuzz(n):\n    # Írd ide a megoldásodat!\n    pass\n",
        "javascript": "function dragonFizzbuzz(n) {\n    // Írd ide a megoldásodat!\n}\n",
        "csharp": "using System;\nusing System.Collections.Generic;\n\npublic class Solution {\n    public static List<string> DragonFizzbuzz(int n) {\n        // Írd ide a megoldásodat!\n        return new List<string>();\n    }\n}\n",
        "php": "<?php\nfunction dragonFizzbuzz($n) {\n    // Írd ide a megoldásodat!\n}\n",
        "go": "package main\n\nfunc DragonFizzbuzz(n int) []string {\n    // Írd ide a megoldásodat!\n    return nil\n}\n"
    }'::jsonb,
    '[
        {"input": "5", "expected": "[\"1\", \"2\", \"Sárkány\", \"4\", \"Mágia\"]"},
        {"input": "15", "expected": "[\"1\", \"2\", \"Sárkány\", \"4\", \"Mágia\", \"Sárkány\", \"7\", \"8\", \"Sárkány\", \"Mágia\", \"11\", \"Sárkány\", \"13\", \"14\", \"Sárkánymágia\"]"}
    ]'::jsonb
),
(
    'valid-anagram',
    'Anagramma Detektor',
    'Döntsd el két szóról (`s` és `t`), hogy egymás anagrammái-e (ugyanazokat a betűket tartalmazzák-e azonos darabszámban)!',
    '💡 **Magyarázat & Megoldási Tippek:**\n\n- **1. Megoldás:** Rendezd mindkét string karaktereit ábécérendbe (pl. `sorted(s) == sorted(t)`). Ha a hosszuk eltér, azonnal `false`.\n- **2. Megoldás:** Számold meg a betűk gyakoriságát egy karaktertáblában.',
    'Könnyű',
    35,
    70,
    '{
        "python": "def is_anagram(s, t):\n    # Írd ide a megoldásodat!\n    pass\n",
        "javascript": "function isAnagram(s, t) {\n    // Írd ide a megoldásodat!\n}\n",
        "csharp": "using System;\n\npublic class Solution {\n    public static bool IsAnagram(string s, string t) {\n        // Írd ide a megoldásodat!\n        return false;\n    }\n}\n",
        "php": "<?php\nfunction isAnagram($s, $t) {\n    // Írd ide a megoldásodat!\n}\n",
        "go": "package main\n\nfunc IsAnagram(s, t string) bool {\n    // Írd ide a megoldásodat!\n    return false\n}\n"
    }'::jsonb,
    '[
        {"input": "\"anagram\", \"nagaram\"", "expected": "true"},
        {"input": "\"rat\", \"car\"", "expected": "false"},
        {"input": "\"kacsa\", \"sakca\"", "expected": "true"}
    ]'::jsonb
),
(
    'longest-word-length',
    'Leghosszabb Szó Hossza',
    'Keresd meg és add vissza a leghosszabb szó karakterszámát a megadott mondatban (`sentence`)!',
    '💡 **Magyarázat & Megoldási Tippek:**\n\n- Vágd szét a mondatot szóközök mentén (pl. `.split()`), majd egy ciklussal vagy maximum kereséssel határozd meg a leghosszabb elem hosszát.',
    'Könnyű',
    30,
    60,
    '{
        "python": "def longest_word_length(sentence):\n    # Írd ide a megoldásodat!\n    pass\n",
        "javascript": "function longestWordLength(sentence) {\n    // Írd ide a megoldásodat!\n}\n",
        "csharp": "using System;\n\npublic class Solution {\n    public static int LongestWordLength(string sentence) {\n        // Írd ide a megoldásodat!\n        return 0;\n    }\n}\n",
        "php": "<?php\nfunction longestWordLength($sentence) {\n    // Írd ide a megoldásodat!\n}\n",
        "go": "package main\n\nfunc LongestWordLength(sentence string) int {\n    // Írd ide a megoldásodat!\n    return 0\n}\n"
    }'::jsonb,
    '[
        {"input": "\"A programozas fantasztikus dolog\"", "expected": "12"},
        {"input": "\"Go Python CSharp PHP\"", "expected": "6"},
        {"input": "\"Egy\"", "expected": "3"}
    ]'::jsonb
)
ON CONFLICT (slug) DO NOTHING;
