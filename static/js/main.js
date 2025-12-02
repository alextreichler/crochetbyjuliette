console.log("Script loaded from static/js/main.js");

document.addEventListener('DOMContentLoaded', function() {
    // Get the modal
    var modal = document.getElementById("imageModal");
    var modalImg = document.getElementById("img01");
    var span = document.getElementsByClassName("close")[0];

    // Get all images in cards
    var images = document.querySelectorAll('.card img');
    console.log("Found " + images.length + " images.");

    if (images.length > 0) {
        images.forEach(function(img) {
            img.style.cursor = "pointer";
            img.onclick = function() {
                console.log("Image clicked: " + this.src);
                modal.style.display = "flex";
                modalImg.src = this.src;
            };
        });
    }

    // When the user clicks on <span> (x), close the modal
    if (span) {
        span.onclick = function() { 
            modal.style.display = "none";
        }
    }

    // When the user clicks anywhere outside of the image, close it
    window.onclick = function(event) {
        if (event.target == modal) {
            modal.style.display = "none";
        }
    // Toast Notification Logic
    var flashMessagesContainer = document.getElementById('toast-target');
    if (flashMessagesContainer) {
        var flashMessages = flashMessagesContainer.querySelectorAll('.flash-message');
        var toastContainer = document.querySelector('.toast-container');
        if (!toastContainer) {
            toastContainer = document.createElement('div');
            toastContainer.className = 'toast-container';
            document.body.appendChild(toastContainer);
        }

        flashMessages.forEach(function(flashMsg) {
            var type = flashMsg.classList.contains('flash-success') ? 'toast-success' : 'toast-error';
            var message = flashMsg.textContent;

            var toast = document.createElement('div');
            toast.className = 'toast ' + type;
            toast.textContent = message;
            toastContainer.appendChild(toast);

            // Animate in
            setTimeout(function() {
                toast.style.opacity = '1';
                toast.style.transform = 'translateX(0)';
            }, 10); // Small delay to ensure CSS transition applies

            // Animate out and remove
            setTimeout(function() {
                toast.style.opacity = '0';
                toast.style.transform = 'translateX(100%)';
                toast.addEventListener('transitionend', function() {
                    toast.remove();
                    // If no more toasts, remove toast-container
                    if (toastContainer.children.length === 0) {
                        toastContainer.remove();
                    }
                });
            }, 5000); // 5 seconds
        });

        // Remove the original flash messages container from the DOM
        flashMessagesContainer.remove();
    }
});
