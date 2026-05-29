async function loadUserStats() {
    try {
        const response = await fetch('/api/user/stats');
        if (!response.ok) throw new Error("Ошибка загрузки статистики");
        
        const data = await response.json();

        // Заполняем счетчики
        document.getElementById('totalPassedDecks').textContent = data.total_passed_decks;
        document.getElementById('overallKnowledge').textContent = `${data.overall_knowledge}%`;

        const listContainer = document.getElementById('passedDecksList');
        
        if (data.passed_decks && data.passed_decks.length > 0) {
            listContainer.innerHTML = ''; // Очищаем текст по умолчанию
            
            data.passed_decks.forEach(deck => {
                const item = document.createElement('div');
                item.className = 'passed-deck-item';
                
                // Проверяем результат для подбора цвета комментария
                const isWarning = deck.percentage < 70;
                const commentClass = isWarning ? 'comment-warn' : 'comment-success';

                item.innerHTML = `
                    <div class="passed-deck-info">
                        <span class="passed-deck-title">${deck.deck_title}</span>
                        <span class="passed-deck-comment ${commentClass}">${deck.comment}</span>
                    </div>
                    <div class="passed-deck-percent">${deck.percentage}%</div>
                `;
                listContainer.appendChild(item);
            });
        }
    } catch (err) {
        console.error("Не удалось подгрузить кузнечную статистику:", err);
    }
}

// Управление аккордеоном (плавное открытие списка)
const toggleBtn = document.getElementById('toggleStatsBtn');
const content = document.getElementById('statsContent');

if (toggleBtn && content) {
    toggleBtn.addEventListener('click', () => {
        const isOpen = content.classList.toggle('open');
        toggleBtn.querySelector('.arrow').style.transform = isOpen ? 'rotate(180deg)' : 'rotate(0deg)';
        
        if (isOpen) {
            content.style.maxHeight = content.scrollHeight + "px";
        } else {
            content.style.maxHeight = "0px";
        }
    });
}

// Запускаем сбор данных при открытии страницы
document.addEventListener('DOMContentLoaded', loadUserStats);