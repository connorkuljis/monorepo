// Function to convert array of objects to CSV
function arrayToCSV(arr) {
  // Initialize CSV content with the header row
  let csvContent = "key,value\n";

  // Iterate over the array and append each object's key-value pair
  arr.forEach((item) => {
    // Escape any double quotes and commas in the value
    const escapedValue = item.value.toString().replace(/"/g, '""');
    csvContent += `"${item.key}","${escapedValue}"\n`;
  });

  return csvContent;
}

function getSeekJobDetails() {
  const map = new Map();
  map.set("job-detail-title", "job");
  map.set("advertiser-name", "company");
  map.set("job-detail-classifications", "industry");
  map.set("company-review", "review");
  map.set("job-detail-apply", "job posting");
  map.set("jobAdDetails", "description");

  const jobDetails = Array.from(document.querySelectorAll("[data-automation]"))
    .filter((el) => map.has(el.getAttribute("data-automation")))
    .map((el) => {
      const attr = el.getAttribute("data-automation");
      const key = map.get(attr);
      let value = el.textContent;

      if (attr === "job-detail-apply") {
        value = el.getAttribute("href");
        value = "https://www.seek.com.au" + value;
      }

      return { key, value };
    });

  // Add timestamp object to the beginning of the array
  return [
    { key: "created-at", value: new Date().toISOString() },
    ...jobDetails,
  ];
}

let ob = getSeekJobDetails();
console.log(ob.description);
// let res = "";

// for (let i = 0; i < ob.length; i++) {
//   if (i > 0) {
//     res += ",";
//   }
//   res += ob[i].value;
// }

function getPrompt() {
  return "Please write a one page cover letter for the given job description. Tailor the cover letter to the supplied resume.";
}
