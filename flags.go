package main

// Challenge holds everything Mission Control needs to render and grade one flag.
type Challenge struct {
	ID          string // stable key, e.g. "f1"
	Number      int
	CodeName    string // in-story name, e.g. "The Front Door"
	Category    string // e.g. "SQL Injection"
	Briefing    string // shown on the case file before solving
	TargetLabel string // human label for where to go, e.g. "Open the Login page"
	TargetURL   string
	Hints       []string // progressive: nudge -> stronger -> near-answer
	Flag        string   // exact string the player must submit
	Reveal      string   // shown after a correct submission - the "real world impact" beat
}

var challenges = []Challenge{
	{
		ID:          "f1",
		Number:      1,
		CodeName:    "The Front Door",
		Category:    "SQL Injection",
		Briefing:    "ShopKart's login form is the very first thing every visitor touches. Our tipster says it's not checking who it lets in as carefully as it should. Can you walk in without a key?",
		TargetLabel: "Open the Login page",
		TargetURL:   "/login",
		Hints: []string{
			"Think about what the login form is really asking the database: \u201Cis there a row where the username AND the password match?\u201D What if you could make that question always true?",
			"Try typing a username that ends the database's sentence early and adds your own condition \u2014 something like:  admin' --  (the -- tells the database to ignore the rest of the line).",
			"Exact payload to try in the Username field:  admin' --   \u2014 leave the password field blank or type anything, then press Login.",
		},
		Flag:   "WIN{a0ca3a0bb29c61f947c8}",
		Reveal: "You just performed a SQL Injection login bypass \u2014 no password required. In a real breach, this is often step one: an attacker walks in as an administrator in under ten seconds, no brute-forcing needed.",
	},
	{
		ID:          "f2",
		Number:      2,
		CodeName:    "The Spare Key",
		Category:    "Broken Authentication",
		Briefing:    "Not every door gets kicked in \u2014 sometimes someone just left a spare key under the mat. Our tipster mentions the team left themselves a note somewhere on the public site. Developers talk to each other in comments. Are you listening?",
		TargetLabel: "Open the About ShopKart page",
		TargetURL:   "/about",
		Hints: []string{
			"Every browser lets you look at the raw HTML behind a page \u2014 right-click anywhere and choose \u201CView Page Source\u201D (or press Ctrl+U / Cmd+Option+U).",
			"Some elements are deliberately hidden from the visible page (so you won't see them just by scrolling) but are still sitting right there in the raw HTML. Scan slowly \u2014 developers sometimes leave notes to each other this way and forget to remove them.",
			"Once you find the leftover credentials in the comment, go to /login and sign in with them directly \u2014 no injection needed this time.",
		},
		Flag:   "WIN{12b332de4e84e4583f20}",
		Reveal: "That's a Broken Authentication finding: a real, working account with a weak, default, or leaked password. No clever exploit required \u2014 just patience and a bit of observation. This is one of the most common ways real breaches actually start.",
	},
	{
		ID:          "f3",
		Number:      3,
		CodeName:    "The Poisoned Review",
		Category:    "Cross-Site Scripting (XSS)",
		Briefing:    "Anyone can leave a product review on ShopKart \u2014 and whatever they type gets shown to every single visitor afterwards, exactly as typed. Our tipster thinks that's a problem. Can you make the page run your own code, not just display your own text?",
		TargetLabel: "Open the Aurora Desk Lamp product page",
		TargetURL:   "/product/1",
		Hints: []string{
			"Normal reviews are just text. But what if the review box doesn't check whether you typed text \u2014 or an actual HTML tag, like <script>?",
			"Try submitting a review whose entire content is exactly:  <script>alert('hi')</script>  \u2014 then look at the page (not the form \u2014 the actual review list below it).",
			"Once you see your own alert box pop up, that proves the page runs your code. Now make it show the page's hidden secret instead of 'hi': submit a review containing exactly  <script>alert(SHOPKART_SECRET)</script>",
		},
		Flag:   "WIN{163f7e4d8deb3cd999ed}",
		Reveal: "That's Stored Cross-Site Scripting. Your review didn't just get displayed \u2014 it got executed, in the browser of anyone who viewed the page. In a real attack this is how session cookies, saved logins, and private data get quietly stolen from other visitors, not just you.",
	},
	{
		ID:          "f4",
		Number:      4,
		CodeName:    "The Leaky API",
		Category:    "Sensitive Data Exposure",
		Briefing:    "Modern websites constantly talk to invisible backend addresses (APIs) to fetch data behind the scenes. Our tipster suspects ShopKart built one of these to list customer accounts \u2014 and forgot to lock the door on it.",
		TargetLabel: "No page for this one \u2014 you'll have to guess the address",
		TargetURL:   "",
		Hints: []string{
			"Websites often expose their data through addresses starting with /api/ \u2014 followed by the name of the thing they're listing.",
			"ShopKart's login page needs a list of usernames to check against somewhere. Try visiting:  /api/users  directly in your browser's address bar.",
			"Once the page loads, look carefully through the JSON \u2014 one of the accounts isn't a real customer, and has a note attached to it that shouldn't be public.",
		},
		Flag:   "WIN{a3902d54465a301968a8}",
		Reveal: "That's a Sensitive Data Exposure finding \u2014 an API endpoint with zero access control, quietly leaking every account's data (including internal notes) to anyone who knows, or guesses, the URL. No login, no hacking tool, just a browser address bar.",
	},
	{
		ID:          "f5",
		Number:      5,
		CodeName:    "The Backdoor Left Open",
		Category:    "Security Misconfiguration",
		Briefing:    "The very last thing before a launch: config files, backups, debug notes \u2014 the paperwork of building a website. All of it is meant to stay on the server, never on the public internet. Our tipster has a bad feeling ShopKart forgot to lock one of these away.",
		TargetLabel: "No page for this one \u2014 config files usually hide in plain sight",
		TargetURL:   "",
		Hints: []string{
			"Almost every web project keeps its secret settings (database passwords, API keys) in one small, predictably-named file.",
			"That file is very often named exactly  .env  and sits at the root of the website \u2014 try:  /.env",
			"If you can read real database passwords sitting right there in your browser, you've found it. Look for a line near the bottom.",
		},
		Flag:   "WIN{4fbecca978479827467f}",
		Reveal: "That's a Security Misconfiguration \u2014 a sensitive file that should never leave the server, sitting in a folder the web server was happy to hand out to anyone who asked. It's one of the most common (and most avoidable) findings in any real assessment.",
	},
}

func challengeByID(id string) *Challenge {
	for i := range challenges {
		if challenges[i].ID == id {
			return &challenges[i]
		}
	}
	return nil
}
