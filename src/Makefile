.PHONY: init run test

init:
	PYTHONPATH=backend python -c "from app import init_db; init_db()"

run:
	python ./backend/app.py

test: 
	PYTHONPATH=backend python ./backend/app_tests.py

