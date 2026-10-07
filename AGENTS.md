This project builds a web invitation platform for events, focusing specifically on 'Quinceañeras', Weddings, and other milestone parties and events.

It is built as a Go monolith, using the built-in `template/html` library. Interactivity can be added where necessary. To keep the codebase maintainable, prefer using tools already included in the project, such as AlpineJS. Avoid writing hundereds of lines of custom CSS or JavaScript, unless absolutely necessary. When possible, use pre-built BulmaCSS components, specially for internal or non-public-facing tools like the `/admin` functionality. The only exception is the invitation styling, which can use CSS to make it aesthetically pleasing. 

All user-facing content you output (such as that displayed in the HTML templates), should use primarily **Spanish language**. However, you don't need to translate any existing English text, those will be corrected eventually.

`PLAN.md` contains the initial project spec.