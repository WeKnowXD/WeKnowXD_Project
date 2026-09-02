setuptools står for sig selv:
Den er ikke en afhængighed af Flask, Werkzeug eller Jinja2 – den er et byggeværktøj, der bruges til selve installations-processen (det er den vi tidligere kæmpede med pga. Feature-fejlen). Den er ikke noget din kode importerer eller bruger ved kørsel, kun noget pip bruger bag kulisserne når den installerer de andre pakker.

Werkzeug – "motorrummet"
Den håndterer alt det lavpraktiske netværks-arbejde: modtager HTTP-requests fra browseren, parser URL'er, håndterer cookies/sessions, og sender HTTP-svar tilbage. Den kører selve web-serveren du så starte i din terminal (Running on http://127.0.0.1:8080/).

Jinja2 – "skabelon-motoren"
Den fylder dynamisk data ind i dine HTML-filer. Når du bruger render_template(...) i din app.py til at vise fx login.html eller about.html, er det Jinja2 der bag kulisserne erstatter ting som {{ brugernavn }} med den rigtige værdi, før HTML'en sendes til browseren.

Flask – "limen imellem dem"
Flask selv indeholder relativt lidt kode – den er primært et bekvemt lag der binder Werkzeug og Jinja2 sammen, og giver dig de simple @app.route(...)-dekoratorer og funktioner du bruger i din kode, i stedet for at du selv skulle konfigurere Werkzeug og Jinja2 manuelt hver gang.
