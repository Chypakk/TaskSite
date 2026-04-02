export {formatDate, getStatusText, getStatusColor, getPriorityText, getPriorityColor};

function formatDate(dateString) {
    if (!dateString || dateString === '0001-01-01T00:00:00Z') return '—';
    const date = new Date(dateString);
    const formatedDate = date.toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
    const splitedDate = formatedDate.split(',');
    return `${splitedDate[1]}, ${splitedDate[0]}`;
}

// Вспомогательные методы
function getStatusText(status) {
    const statuses = {
        'open': 'Открыта',
        'in_progress': 'В работе',
        'completed': 'Завершена',
        'closed': 'Отменена'
    };
    return statuses[status] || status;
}

function getStatusColor(status) {
    const colors = {
        'open': 'warning',
        'in_progress': 'primary',
        'completed': 'success',
        'closed': 'secondary'
    };
    return colors[status] || 'secondary';
}

function getPriorityText(priority) {
    const priorities = {
        'low': 'Низкий',
        'medium': 'Средний',
        'high': 'Высокий',
        'critical': 'Критический'
    };
    return priorities[priority] || priority;
}

function getPriorityColor(priority) {
    const colors = {
        'low': 'secondary',
        'medium': 'primary',
        'high': 'warning',
        'critical': 'danger'
    };
    return colors[priority] || 'secondary';
}