jQuery.fn.extend({
  dataBind: function (name) {
    return this.find('[data-bind=' + name + ']');
  }
});

var ROOT_URL = window.location.href;

var $listMail = $('#list-mail'),
    $header = $('header.mail-header');

var header = {
  to: $header.dataBind('to'),
  from: $header.dataBind('from'),
  subject: $header.dataBind('subject'),
  date: $header.dataBind('date')
};


function fetchMailList(){
  $.get(ROOT_URL + 'api/mail').done(function(mails) {
    $listMail.html('');
    for(i = mails.length -1; i >= 0; i--){
      addMailItem($listMail, mails[i]);
    }
    // Apply material effects
    $.material.init();
  })
}

function displayMail(id){
  $.get(ROOT_URL + 'api/mail/' + id).done(function(mail) {
    mail = parseMailMeta(mail, id);

    header.to.text(mail.To);
    header.from.text(mail.From);
    header.subject.text(mail.Subject);
    header.date.text(mail.Date);

  });

  $this = $('#list-mail').find('[data-id=' + id + ']');

  $this.attr('data-read', true);

  $listMail.children().removeClass('active');
  $this.addClass('active');

  // Display the mail in our iframe
  $('#mail-body').attr('src', ROOT_URL + '/api/mail/raw/' + id);
}

function setupNotifications() {
  // Notify the user if notifications are not supported
  if(!('Notification' in window)){
    alert('Sadly, your browser does not support notifications');
  }

  // We don't have permission yet, request it
  else if(Notification.permission == 'denied' || Notification.permission === 'default'){
    Notification.requestPermission().then(function (result) {
      if (result === 'denied') {
        alert('Please allow us to send you notifications');
      }
    })
  }
}

function sendMailNotification(mail){
  // Only continue if we are allowed to send notifications
  if (Notification.permission !== 'granted') {
    return;
  }

  var notify = new Notification('New mail from "' + mail.From + '"', {
    body: mail.Subject,
    icon : ROOT_URL + 'asset/image/envelope.svg'
  });

  notify.addEventListener('click', function () {
    displayMail(mail.Id);
  });
}

function setupWebSocket() {
  // Create a websocket
  var socket = new WebSocket("ws://localhost:8025/websocket", "protocolOne");
  socket.onopen = function(){
    // We now have a active connection to the daemon
  };

  socket.addEventListener('close', function() {
    alert('Socket connection closed, please reload');
  });

  // We receive a message if Mail Dev has received a new e-mail, so lets fetch it
  socket.addEventListener('message', function (event) {

    var mailId = event.data;
    $.get(ROOT_URL + 'api/mail/' + mailId).done(function(mail){
      mail =  parseMailMeta(mail, mailId);

      // Add the mail to our list
      addMailItem($listMail, mail);

      // Send a notification to inform the user about this mail
      sendMailNotification(mail);

    }).error(function() {
      alert('An error occurred while fetching a new e-mail, please reload');
    });
  });
}

function addMailItem($listMail, mail){
  if(mail.Id == ''){
    return;
  }

  date = new Date(mail.Date);
  $listMail.prepend(
      '<li class="withripple mail-item" data-id="' + mail.Id + '" data-read="false">' +
      '<div class="top"><strong>' + mail.To + '</strong><span>' + date.toLocaleTimeString() + '</span></div>' +
      mail.Subject +
      '</li>');

  $.material.init();
}

function parseMailMeta(mail, id){

  var from = '{nobody}',
      to = '{nobody}',
      subject = '{No subject}',
      date = '{No date}';

  if(mail.Headers.From[0]){
    from = mail.Headers.From[0]
  }

  if(mail.Headers.To[0]){
    to = mail.Headers.To[0]
  }

  if(mail.Headers.Subject[0]){
    subject = mail.Headers.Subject[0]
  }

  if(mail.Headers.Date[0]){
    date = mail.Headers.Date[0]
  }

  return {
    Id: id,
    From: from,
    To: to,
    Subject : subject,
    Date : date
  };
}

// FIXME: We should fix this with only CSS3
$(window).on("resize", function () {
  $("html, body").height($(window).height());
  $(".main, .menu").height($(window).height() - $(".header-panel").outerHeight());
  $(".pages").height($(window).height());
}).trigger("resize");

$(document).ready(function(){
  setupWebSocket();
  setupNotifications();
  fetchMailList();

  $listMail.on('click', '.mail-item', function(){
    var $this = $(this),
        mailId = $this.data('id');

    displayMail(mailId);
  });
});