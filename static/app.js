"use strict";

const SAMPLE_BACKGROUND_URL =
  "https://images.unsplash.com/photo-1570246159995-57eaeeca884b?q=80&w=1288&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D";

const GERMAN_MONTHS = [
  "Januar", "Februar", "März", "April", "Mai", "Juni",
  "Juli", "August", "September", "Oktober", "November", "Dezember",
];
const GERMAN_WEEKDAYS_SHORT = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];
const GERMAN_WEEKDAYS_LONG = [
  "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag",
];

const state = {
  config: { locationLabel: "Zuhause", clockFormat: "24h", accentColor: "#F5A524", showIndoor: false },
  bgMonthChecked: null,
};

const els = {
  canvas: document.getElementById("canvas"),
  bgPhoto: document.getElementById("bg-photo"),
  clock: document.getElementById("clock"),
  date: document.getElementById("date"),
  temp: document.getElementById("temp"),
  icon: document.getElementById("icon"),
  location: document.getElementById("location"),
  feels: document.getElementById("feels"),
  sunrise: document.getElementById("sunrise"),
  sunset: document.getElementById("sunset"),
  indoorStat: document.getElementById("indoor-stat"),
  indoor: document.getElementById("indoor"),
  forecast: document.getElementById("forecast"),
  weekdayRow: document.getElementById("weekday-row"),
  calendarGrid: document.getElementById("calendar-grid"),
  alertBanner: document.getElementById("alert-banner"),
  alertText: document.getElementById("alert-text"),
  joke: document.getElementById("joke"),
};

// ---------- viewport scaling ----------
// The design is a fixed 1080x1920 canvas; scale it to fit whatever the
// actual TV/browser viewport is while preserving proportions.
function applyScale() {
  const scale = Math.min(window.innerWidth / 1080, window.innerHeight / 1920);
  els.canvas.style.transform = `scale(${scale})`;
}
window.addEventListener("resize", applyScale);
applyScale();

// ---------- weather icons ----------

function weatherIconSVG(code, size) {
  size = size || 56;
  const c = code ?? 0;
  let key = "sun";
  if (c === 0) key = "sun";
  else if (c === 1 || c === 2) key = "cloudsun";
  else if (c === 3) key = "cloud";
  else if (c === 45 || c === 48) key = "fog";
  else if ([51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 80, 81, 82].includes(c)) key = "rain";
  else if ([71, 73, 75, 77, 85, 86].includes(c)) key = "snow";
  else if ([95, 96, 99].includes(c)) key = "storm";

  const sunFill = "#FFC94D";
  const cloudFill = "#C9D6E8";
  const head = `<svg width="${size}" height="${size}" viewBox="0 0 48 48" fill="none">`;
  const tail = `</svg>`;

  if (key === "sun") {
    const rays = [0, 45, 90, 135, 180, 225, 270, 315].map((a) => {
      const rad = (a * Math.PI) / 180;
      const x1 = 24 + Math.cos(rad) * 15, y1 = 24 + Math.sin(rad) * 15;
      const x2 = 24 + Math.cos(rad) * 20, y2 = 24 + Math.sin(rad) * 20;
      return `<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}"/>`;
    }).join("");
    return `${head}<circle cx="24" cy="24" r="10" fill="${sunFill}"/><g stroke="${sunFill}" stroke-width="2.4" stroke-linecap="round">${rays}</g>${tail}`;
  }
  if (key === "cloudsun") {
    return `${head}<circle cx="18" cy="18" r="8" fill="${sunFill}"/><path d="M10 32a8 8 0 0 1 6-14 10 10 0 0 1 19 3 7 7 0 0 1-1 21H16a6 6 0 0 1-6-10z" fill="${cloudFill}"/>${tail}`;
  }
  if (key === "cloud") {
    return `${head}<path d="M10 32a8 8 0 0 1 6-14 10 10 0 0 1 19 3 7 7 0 0 1-1 21H16a6 6 0 0 1-6-10z" fill="${cloudFill}"/>${tail}`;
  }
  if (key === "fog") {
    return `${head}<path d="M12 20a8 8 0 0 1 6-13 10 10 0 0 1 18 4 7 7 0 0 1-1 19H18a6 6 0 0 1-6-10z" fill="#B9C6D8"/><g stroke="#9FB0C6" stroke-width="2.4" stroke-linecap="round"><line x1="10" y1="34" x2="38" y2="34"/><line x1="10" y1="40" x2="38" y2="40"/></g>${tail}`;
  }
  if (key === "rain") {
    return `${head}<path d="M10 24a8 8 0 0 1 6-14 10 10 0 0 1 19 3 7 7 0 0 1-1 21H16a6 6 0 0 1-6-10z" fill="${cloudFill}"/><g stroke="#6FA8FF" stroke-width="2.6" stroke-linecap="round"><line x1="16" y1="36" x2="13" y2="44"/><line x1="24" y1="36" x2="21" y2="44"/><line x1="32" y1="36" x2="29" y2="44"/></g>${tail}`;
  }
  if (key === "snow") {
    return `${head}<path d="M10 24a8 8 0 0 1 6-14 10 10 0 0 1 19 3 7 7 0 0 1-1 21H16a6 6 0 0 1-6-10z" fill="${cloudFill}"/><g stroke="#EAF2FF" stroke-width="2.4" stroke-linecap="round"><line x1="16" y1="35" x2="16" y2="45"/><line x1="24" y1="35" x2="24" y2="45"/><line x1="32" y1="35" x2="32" y2="45"/></g>${tail}`;
  }
  // storm
  return `${head}<path d="M10 22a8 8 0 0 1 6-14 10 10 0 0 1 19 3 7 7 0 0 1-1 21H16a6 6 0 0 1-6-10z" fill="#A9B7CC"/><path d="M25 32l-6 9h5l-3 7 9-11h-5l3-5z" fill="${sunFill}"/>${tail}`;
}

// ---------- date helpers ----------

function dateKey(d) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function formatTime(now, format) {
  const h24 = String(now.getHours()).padStart(2, "0");
  const m = String(now.getMinutes()).padStart(2, "0");
  if (format === "12h") {
    let h = now.getHours() % 12;
    if (h === 0) h = 12;
    const ampm = now.getHours() < 12 ? "AM" : "PM";
    return `${h}:${m} ${ampm}`;
  }
  return `${h24}:${m}`;
}

// ---------- clock + background ----------

function renderClock() {
  const now = new Date();
  els.clock.textContent = formatTime(now, state.config.clockFormat);
  const weekdayIdx = (now.getDay() + 6) % 7;
  els.date.textContent = `${GERMAN_WEEKDAYS_LONG[weekdayIdx]}, ${now.getDate()}. ${GERMAN_MONTHS[now.getMonth()]}`;

  const monthIdx = now.getMonth() + 1;
  if (state.bgMonthChecked !== monthIdx) {
    checkBackground();
  }
}

function checkBackground() {
  const monthIdx = new Date().getMonth() + 1;
  const src = `backgrounds/${String(monthIdx).padStart(2, "0")}.jpg`;
  const img = new Image();
  img.onload = () => {
    els.bgPhoto.src = src;
    state.bgMonthChecked = monthIdx;
  };
  img.onerror = () => {
    els.bgPhoto.src = SAMPLE_BACKGROUND_URL;
    state.bgMonthChecked = monthIdx;
  };
  img.src = src;
}

// ---------- weather ----------

async function fetchWeather() {
  try {
    const res = await fetch("/api/weather");
    const data = await res.json();
    renderWeather(data);
  } catch (e) {
    renderWeather(null);
  }
}

function renderWeather(data) {
  const cur = data && data.current;
  const daily = (data && data.daily) || [];
  const accent = state.config.accentColor;

  els.temp.textContent = cur ? `${Math.round(cur.temperature)}°` : "—";
  els.icon.innerHTML = weatherIconSVG(cur ? cur.weatherCode : 0, 90);
  els.location.textContent = state.config.locationLabel;
  els.feels.textContent = `Gefühlt wie ${cur ? Math.round(cur.feelsLike) : "—"}°`;

  const today = daily[0];
  els.sunrise.textContent = (today && today.sunrise) || "—";
  els.sunset.textContent = (today && today.sunset) || "—";

  els.forecast.innerHTML = "";
  daily.slice(0, 6).forEach((day, i) => {
    const d = new Date(day.date + "T00:00:00");
    const isToday = i === 0;
    const label = isToday ? "Heute" : GERMAN_WEEKDAYS_SHORT[(d.getDay() + 6) % 7];
    const col = document.createElement("div");
    col.className = "forecast-day";
    col.innerHTML = `
      <div class="label${isToday ? " today" : ""}" ${isToday ? `style="color:${accent}"` : ""}>${label}</div>
      <div class="icon">${weatherIconSVG(day.weatherCode, 44)}</div>
      <div class="pop">${day.pop ?? 0}%</div>
      <div class="tmax">${Math.round(day.tempMax)}°</div>
      <div class="tmin">${Math.round(day.tempMin)}°</div>
    `;
    els.forecast.appendChild(col);
  });
}

// ---------- indoor sensor ----------

async function fetchIndoor() {
  if (!state.config.showIndoor) return;
  try {
    const res = await fetch("/api/indoor");
    const data = await res.json();
    renderIndoor(data);
  } catch (e) {
    renderIndoor(null);
  }
}

function renderIndoor(data) {
  if (!state.config.showIndoor) return;
  els.indoorStat.hidden = false;
  const temp = data && data.temperature;
  const hum = data && data.humidity;
  let text = "—";
  if (temp != null) {
    text = `${Math.round(temp)}°`;
    if (hum != null) text += ` / ${Math.round(hum)}%`;
  }
  els.indoor.textContent = text;
}

// ---------- joke ----------

async function fetchJoke() {
  try {
    const res = await fetch("/api/joke");
    const data = await res.json();
    renderJoke(data && data.text);
  } catch (e) {
    renderJoke(null);
  }
}

function renderJoke(text) {
  if (!text) {
    els.joke.hidden = true;
    return;
  }
  els.joke.hidden = false;
  els.joke.textContent = `"${text}"`;
}

// ---------- alerts ----------

const SEVERITY_RANK = { Extreme: 3, Severe: 3, Moderate: 2, Minor: 1 };

async function fetchAlerts() {
  try {
    const res = await fetch("/api/alerts");
    const data = await res.json();
    renderAlert(data && data.alert);
  } catch (e) {
    renderAlert(null);
  }
}

function renderAlert(alert) {
  if (!alert) {
    els.alertBanner.hidden = true;
    return;
  }
  const isSevere = (SEVERITY_RANK[alert.severity] || 0) >= 3;
  const color = isSevere ? "#FF5A5F" : "#F5A524";
  els.alertBanner.hidden = false;
  els.alertBanner.style.background = `${color}26`;
  els.alertBanner.style.border = `1px solid ${color}88`;
  els.alertBanner.style.color = color;
  els.alertText.textContent = alert.headline;
}

// ---------- calendar ----------

let weekdayRowRendered = false;

function renderWeekdayRow() {
  if (weekdayRowRendered) return;
  els.weekdayRow.innerHTML = GERMAN_WEEKDAYS_SHORT.map((wd) => `<div class="wd">${wd}</div>`).join("");
  weekdayRowRendered = true;
}

let lastEvents = [];

async function fetchCalendar() {
  try {
    const res = await fetch("/api/calendar");
    const data = await res.json();
    lastEvents = (data && data.events) || [];
  } catch (e) {
    lastEvents = [];
  }
  renderCalendar();
}

function renderCalendar() {
  renderWeekdayRow();
  const accent = state.config.accentColor;
  const now = new Date();
  const todayMidnight = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const curWeekday = (todayMidnight.getDay() + 6) % 7; // 0 = Monday
  const gridStart = new Date(todayMidnight);
  gridStart.setDate(gridStart.getDate() - curWeekday);

  const eventsByDate = {};
  for (const ev of lastEvents) {
    (eventsByDate[ev.date] = eventsByDate[ev.date] || []).push(ev);
  }

  const todayKey = dateKey(now);
  const maxShown = 4;
  els.calendarGrid.innerHTML = "";

  for (let i = 0; i < 28; i++) {
    const d = new Date(gridStart);
    d.setDate(d.getDate() + i);
    const key = dateKey(d);
    const isToday = key === todayKey;
    const isPast = d < todayMidnight;
    const isOtherMonth = d.getMonth() !== todayMidnight.getMonth();
    const opacity = isPast && !isToday ? 0.4 : isOtherMonth ? 0.55 : 1;

    const dayEvents = (eventsByDate[key] || []).slice().sort((a, b) =>
      a.isWaste === b.isWaste ? 0 : a.isWaste ? -1 : 1
    );
    const visible = dayEvents.slice(0, maxShown);
    const overflow = dayEvents.length > maxShown ? dayEvents.length - maxShown : 0;

    const cell = document.createElement("div");
    cell.className = "cal-cell";
    cell.style.opacity = opacity;

    const dayNum = document.createElement("div");
    dayNum.className = "day-num" + (isToday ? " today" : "");
    dayNum.textContent = d.getDate();
    cell.appendChild(dayNum);

    for (const ev of visible) {
      const pill = document.createElement("div");
      pill.className = "pill " + (ev.isWaste ? "waste" : "personal");
      pill.textContent = ev.title;
      if (ev.isWaste) {
        pill.style.background = `${accent}33`;
        pill.style.color = accent;
        pill.style.border = `1px solid ${accent}77`;
      }
      cell.appendChild(pill);
    }

    if (overflow > 0) {
      const more = document.createElement("div");
      more.className = "overflow";
      more.textContent = `+${overflow} mehr`;
      cell.appendChild(more);
    }

    els.calendarGrid.appendChild(cell);
  }
}

// ---------- init ----------

async function loadConfig() {
  try {
    const res = await fetch("/api/config");
    const data = await res.json();
    state.config = { ...state.config, ...data };
  } catch (e) {
    // keep defaults
  }
  document.documentElement.style.setProperty("--accent", state.config.accentColor);
}

async function init() {
  await loadConfig();

  checkBackground();
  renderClock();
  setInterval(renderClock, 30 * 1000);

  fetchWeather();
  setInterval(fetchWeather, 15 * 60 * 1000);

  fetchJoke();
  setInterval(fetchJoke, 60 * 60 * 1000);

  fetchAlerts();
  setInterval(fetchAlerts, 5 * 60 * 1000);

  fetchCalendar();
  setInterval(fetchCalendar, 15 * 60 * 1000);

  if (state.config.showIndoor) {
    fetchIndoor();
    setInterval(fetchIndoor, 60 * 1000);
  }

  // Recompute the rolling calendar grid at midnight even if the calendar
  // data itself hasn't changed, since "today" shifted.
  setInterval(renderCalendar, 60 * 1000);
}

init();
