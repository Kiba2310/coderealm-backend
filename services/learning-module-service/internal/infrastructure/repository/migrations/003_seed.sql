INSERT INTO courses (id, title, description) VALUES
('python-basics', 'Learn Python', 'Welcome to "Learn Python"! In this course, you''ll learn how to program in Python, one of the most popular and versatile programming languages.')
ON CONFLICT (id) DO NOTHING;

INSERT INTO chapters (id, course_id, order_num, title) VALUES
('ch1-variables', 'python-basics', 1, 'CH1: Variables'),
('ch2-functions', 'python-basics', 2, 'CH2: Functions and Logic')
ON CONFLICT (id) DO NOTHING;

INSERT INTO lessons (id, chapter_id, order_num, title, description, type, language, starter_code, solution_code, test_code) VALUES
(
    'l1-intro',
    'ch1-variables',
    1,
    'L1: Learn Python',
    E'During this course, we''ll be building parts of a make-believe software project: Textio! A SaaS app for sending SMS messages to customers.\n\nEdit the code to print `starting server...` to the console instead of `edit me`.',
    'code_challenge',
    'python',
    'print("edit me")',
    'print("starting server...")',
    ''
),
(
    'l2-variables',
    'ch1-variables',
    2,
    'L2: Variables',
    'In Python, you declare variables by simply assigning a value to a name. Which of the following is a valid variable name?',
    'multiple_choice',
    NULL,
    NULL,
    NULL,
    NULL
),
(
    'l3-string',
    'ch1-variables',
    3,
    'L3: Strings',
    'Strings in Python are sequences of characters. Declare a variable called `greeting` that holds the value `"Hello, World!"`.',
    'code_challenge',
    'python',
    E'# Declare your variable here\n',
    'greeting = "Hello, World!"',
    E'\nimport sys\ntry:\n    if greeting == "Hello, World!":\n        sys.exit(0)\n    else:\n        print("Greeting variable is not correct")\n        sys.exit(1)\nexcept NameError:\n    print("greeting variable is not defined")\n    sys.exit(1)'
),
(
    'l4-theory-loops',
    'ch2-functions',
    1,
    'L4: Loops',
    'A loop is used for iterating over a sequence.',
    'theory',
    NULL,
    NULL,
    NULL,
    NULL
),
(
    'l5-quiz-def',
    'ch2-functions',
    2,
    'L6: Defining functions',
    'Which keyword is used to define a function in Python?',
    'multiple_choice',
    NULL,
    NULL,
    NULL,
    NULL
),
(
    'l6-code-factorial',
    'ch2-functions',
    3,
    'L7: Factorial',
    'Write a `factorial(n)` function.',
    'code_challenge',
    'python',
    E'def factorial(n):\n    # Write the function body here\n    pass\n',
    E'def factorial(n):\n    if n == 0:\n        return 1\n    return n * factorial(n - 1)',
    E'import sys\ntry:\n    assert factorial(5) == 120, "Error: factorial(5) is not 120!"\n    assert factorial(0) == 1, "Error: factorial(0) is not 1!"\n    print("Success!")\nexcept AssertionError as e:\n    print(e)\n    sys.exit(1)\nexcept Exception as e:\n    print("Unknown error:", e)\n    sys.exit(1)'
),
(
    'l7-comp-dict',
    'ch2-functions',
    4,
    'L8: Dictionaries',
    'Complete the code to print the age of the person.',
    'code_challenge',
    'python',
    E'person = {\n    "name": "John",\n    "age": 30,\n    "city": "New York"\n}\n\n# Fill in the blanks!\nx = person[___]\nprint(x)',
    E'person = {\n    "name": "John",\n    "age": 30,\n    "city": "New York"\n}\n\nx = person["age"]\nprint(x)',
    E'import sys\n'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO answer_options (id, lesson_id, text, is_correct) VALUES
('opt1', 'l2-variables', '1st_variable', false),
('opt2', 'l2-variables', 'my_var',       true),
('opt3', 'l2-variables', 'my-var',       false),
('opt4', 'l2-variables', 'class',        false),
('opt-py-new-1', 'l5-quiz-def', 'func', false),
('opt-py-new-2', 'l5-quiz-def', 'def', true),
('opt-py-new-3', 'l5-quiz-def', 'function', false),
('opt-py-new-4', 'l5-quiz-def', 'define', false)
ON CONFLICT (id) DO NOTHING;
