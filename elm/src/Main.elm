module Main exposing (..)

import Array exposing (Array)
import Browser
import Dict exposing (Dict)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Url exposing (percentEncode)



-- MAIN


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }



-- MODEL


type alias Meme =
    { id : String
    , name : String
    , filename : String
    , font_color : String
    , line_chars : Int
    , boxes : List (List Int)
    }


defaultMeme : Meme
defaultMeme =
    Meme "" "" "" "" 0 []


type State
    = Failure
    | Loading
    | Success


type alias Model =
    { memes : Dict String Meme
    , selectedMeme : String
    , texts : Array String
    , state : State
    }


defaultModel =
    Model Dict.empty "" Array.empty Loading


init : () -> ( Model, Cmd Msg )
init _ =
    ( defaultModel, getConfig )



-- UPDATE


type Msg
    = RefreshConfig
    | GotConfig (Result Http.Error (Dict String Meme))
    | MemeChanged String
    | TextChanged Int String


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RefreshConfig ->
            ( defaultModel
            , getConfig
            )

        MemeChanged id ->
            let
                meme =
                    Maybe.withDefault defaultMeme (Dict.get id model.memes)

                numTexts =
                    List.length meme.boxes

                savedText n =
                    Maybe.withDefault "" (Array.get n model.texts)
            in
            ( { model | selectedMeme = id, texts = Array.initialize numTexts savedText }
            , Cmd.none
            )

        GotConfig result ->
            case result of
                Ok memes ->
                    let
                        firstMeme =
                            Maybe.withDefault defaultMeme (List.head (Dict.values memes))

                        numTexts =
                            List.length firstMeme.boxes
                    in
                    ( Model memes firstMeme.id (Array.repeat numTexts "") Success
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | state = Failure }
                    , Cmd.none
                    )

        TextChanged index str ->
            ( { model | texts = Array.set index str model.texts }
            , Cmd.none
            )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> Html Msg
view model =
    case model.state of
        Failure ->
            div []
                [ text "Não consegui carregar a configuração inicial. Tente novamente. "
                , button [ onClick RefreshConfig ] [ text "De novo!" ]
                ]

        Loading ->
            text "Carregando..."

        Success ->
            div []
                [ h1 [] [ text "Escolha o Meme:" ]
                , div [ class "row" ]
                    [ div [ class "col-md-4 mb-4 mb-md-0" ]
                        [ select [ value model.selectedMeme, class "mb-4 form-control", onInput MemeChanged ]
                            [ optgroup [ attribute "label" "Implementados" ]
                                (List.map viewMeme (Dict.values model.memes))
                            ]
                        , fieldset [] (Array.toIndexedList model.texts |> List.map createText)
                        ]
                    , div [ class "col-md-8" ]
                        [ img
                            [ class "mw-100"
                            , src ("/meme?meme=" ++ model.selectedMeme ++ "&" ++ encodeTexts model.texts)
                            ]
                            []
                        ]
                    ]
                ]


viewMeme : Meme -> Html Msg
viewMeme meme =
    option [ value meme.id ] [ text meme.name ]


createText : ( Int, String ) -> Html Msg
createText ( index, str ) =
    input
        [ onInput (TextChanged index)
        , class "form-control"
        , type_ "text"
        , value str
        ]
        []


encodeTexts : Array String -> String
encodeTexts texts =
    let
        encodeParam param =
            percentEncode "text[]" ++ "=" ++ percentEncode param
    in
    Array.map encodeParam texts
        |> Array.toList
        |> String.join "&"



-- HTTP


getConfig : Cmd Msg
getConfig =
    Http.get
        { url = "/config.json"
        , expect = Http.expectJson GotConfig configDecoder
        }



-- JSON decode


configDecoder : D.Decoder (Dict String Meme)
configDecoder =
    D.dict memeDecoder


memeDecoder : D.Decoder Meme
memeDecoder =
    D.map6 Meme
        (D.field "id" D.string)
        (D.field "name" D.string)
        (D.field "filename" D.string)
        (D.field "font-color" D.string)
        (D.field "line-chars" D.int)
        (D.field "boxes" (D.list (D.list D.int)))
