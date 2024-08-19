const ad = document.querySelector(
  '[data-automation="jobAdDetails"]',
).textContent;

fetch("http://localhost:6969/gen", {
  method: "POST",
  headers: {
    "Content-Type": "application/json; charset=UTF-8",
  },
  body: JSON.stringify({
    description: ad,
  }),
})
  .then((response) => {
    if (!response.ok) {
      throw new Error(`Error: ${response.status}`);
    }
    return response.text();
  })
  .then((text) => console.log(text))
  .catch((error) => console.error(error));
