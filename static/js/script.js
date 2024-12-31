document.addEventListener("DOMContentLoaded", function () {
    const form = document.querySelector('form');
    const responseDiv = document.querySelector('#response');

    form.onsubmit = async (e) => {
        e.preventDefault(); // Prevent the default form submit
        const formData = new FormData(form);

        try {
            const res = await fetch('/compress', {
                method: 'POST',
                body: formData,
            });

            const data = await res.json();

            if (res.ok) {
                responseDiv.innerHTML = `<p>${data.message}</p>`;
                if (data.file) {
                    responseDiv.innerHTML += `<a href="${data.file}" download>Download Compressed File</a>`;
                }
            } else {
                responseDiv.innerHTML = `<p style="color: red;">${data.error}</p>`;
            }
        } catch (error) {
            responseDiv.innerHTML = `<p style="color: red;">Error: ${error.message}</p>`;
        }
    };
});
