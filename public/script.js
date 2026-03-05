document.addEventListener('DOMContentLoaded', () => {
    const loader = document.getElementById('loader');
    const exerciseGrid = document.getElementById('exercise-grid');
    const errorMessage = document.getElementById('error-message');

    // Fetch exercises from the updated Go backend endpoint
    fetch('/api/exercises')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            // Hide loader
            loader.classList.add('hidden');

            // Limit to 50 items so the browser doesn't lag rendering thousands of cards at once
            const displayData = data.slice(0, 50);

            // Generate cards
            displayData.forEach(exercise => {
                const card = createExerciseCard(exercise);
                exerciseGrid.appendChild(card);
            });

            // Show grid
            exerciseGrid.classList.remove('hidden');

            // Add subtle entrance animation with a staggered delay
            const cards = document.querySelectorAll('.exercise-card');
            cards.forEach((card, index) => {
                card.style.opacity = '0';
                card.style.transform = 'translateY(20px)';

                setTimeout(() => {
                    card.style.transition = 'all 0.5s cubic-bezier(0.2, 0.8, 0.2, 1)';
                    card.style.opacity = '1';
                    card.style.transform = 'translateY(0)';
                }, 50 * index); // Stagger by 50ms
            });
        })
        .catch(error => {
            console.error('Error fetching data:', error);
            loader.classList.add('hidden');
            errorMessage.classList.remove('hidden');
        });

    function createExerciseCard(exercise) {
        const card = document.createElement('div');
        card.className = 'exercise-card';

        // Grab primary muscle if it exists
        const primaryMuscle = exercise.primaryMuscles && exercise.primaryMuscles.length > 0
            ? exercise.primaryMuscles[0]
            : 'Compound';

        // Grab first string of instructions if exists
        const instructionSnippet = exercise.instructions && exercise.instructions.length > 0
            ? exercise.instructions[0]
            : 'Complete the exercise with proper form.';

        // Image Handling (We fetch the raw images directly from the official database's GitHub repo)
        let imageHtml = '';
        if (exercise.images && exercise.images.length > 0) {
            const imageUrl = `https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/exercises/${exercise.images[0]}`;
            imageHtml = `
                <div class="card-image-container">
                    <img src="${imageUrl}" alt="${exercise.name}" class="card-image" loading="lazy" onerror="this.parentElement.innerHTML='<div class=\\'card-image-fallback\\'>No Image Available</div>'">
                </div>
            `;
        } else {
            imageHtml = `
                <div class="card-image-container">
                    <div class="card-image-fallback">No Image Available</div>
                </div>
            `;
        }

        card.innerHTML = `
            ${imageHtml}
            <div class="card-content">
                <div class="card-title">${exercise.name}</div>
                <div class="card-muscle">${primaryMuscle}</div>
                <div class="card-details">
                    ${exercise.level ? `<span class="badge">${exercise.level}</span>` : ''}
                    ${exercise.equipment ? `<span class="badge">${exercise.equipment}</span>` : ''}
                    ${exercise.category ? `<span class="badge">${exercise.category}</span>` : ''}
                </div>
                <div class="card-instructions">
                    ${instructionSnippet}
                </div>
            </div>
        `;

        return card;
    }
});
